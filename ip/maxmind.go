package ip

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5" // #nosec
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/labstack/gommon/log"
	"github.com/oschwald/geoip2-golang"
)

const maxmindDir = "./resources/maxmind/"

var maxmindFiles = []struct {
	url  string
	md5  string
	file string
	t    string
}{
	{
		url:  "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&suffix=tar.gz",
		md5:  "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&suffix=tar.gz.md5",
		file: "GeoLite2-ASN.mmdb",
		t:    "asn",
	},
	{
		url:  "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&suffix=tar.gz",
		md5:  "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&suffix=tar.gz.md5",
		file: "GeoLite2-Country.mmdb",
		t:    "country",
	},
}

type maxmind struct {
	license string
	country *geoip2.Reader
	asn     *geoip2.Reader
}

func newMaxmind(license string) *maxmind {
	if license == "" {
		log.Infof("[geoip] MaxMind license is not present, reputation accuracy is degraded.")
	}

	return &maxmind{
		license: license,
	}
}

func (g *maxmind) getCountry(ip net.IP) (string, error) {
	if g.country == nil {
		return "-", nil
	}

	country, err := g.country.Country(ip)
	if err != nil {
		return "-", fmt.Errorf("cannot get city for: %s , error: %w", ip, err)
	}

	log.Debugf("[geoip] IP: %s country: %s", ip, country.Country.IsoCode)
	return country.Country.IsoCode, nil
}

func (g *maxmind) getCompany(ip net.IP) (string, error) {
	if g.asn == nil {
		return "-", nil
	}

	asn, err := g.asn.ASN(ip)
	if err != nil {
		return "-", fmt.Errorf("cannot get ASN for: %s , error: %w", ip, err)
	}

	log.Debugf("[geoip] IP: %s ASN: %s", ip, asn.AutonomousSystemOrganization)
	return asn.AutonomousSystemOrganization, nil
}

func (g *maxmind) update() error {
	log.Debug("[geoip] update start")
	defer log.Debug("[geoip] update finished")

	if err := os.MkdirAll(maxmindDir, 0750); err != nil {
		return fmt.Errorf("cannot create directory %s , error: %w", maxmindDir, err)
	}

	for _, m := range maxmindFiles {
		license := "&license_key=" + g.license

		if err := g.download(m.url+license, m.md5+license); err != nil {
			return err
		}

		dbFile := filepath.Join(maxmindDir, m.file)
		db, err := geoip2.Open(dbFile)
		if err != nil {
			return fmt.Errorf("cannot open maxmind file %s, error: %w", dbFile, err)
		}

		if m.t == "country" {
			g.country = db
		} else if m.t == "asn" {
			g.asn = db
		} else {
			return errors.New("invalid type")
		}
	}
	return nil
}

func (g *maxmind) download(url string, md5Url string) error {
	response, err := defaultHTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("cannot download url: %s, error: %w", url, err)
	}
	defer response.Body.Close()

	var md5FileResponse bytes.Buffer
	body := io.TeeReader(response.Body, &md5FileResponse)

	gzr, err := gzip.NewReader(body)
	if err != nil {
		return fmt.Errorf("cannot open GZIP reader url: %s, error: %w", url, err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error while reading from TAR url: %s , error: %w", url, err)
		}

		if header.Typeflag == tar.TypeReg {
			target := filepath.Join(maxmindDir, header.FileInfo().Name())

			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("cannot create file %s, error: %w", target, err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("cannot copy to file %s, error: %w", target, err)
			}

			if err := f.Close(); err != nil {
				return fmt.Errorf("cannot copy close file %s, error: %w", target, err)
			}
		}
	}

	response, err = defaultHTTPClient.Get(md5Url)
	if err != nil {
		return fmt.Errorf("cannot download url: %s, error: %w", md5Url, err)
	}
	defer response.Body.Close()

	md5CheckResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("can read md5 sum from url: %s, error: %w", md5Url, err)
	}

	md5File := fmt.Sprintf("%x", md5.Sum(md5FileResponse.Bytes())) // #nosec
	md5Checksum := string(md5CheckResponse)
	if md5Checksum != md5File {
		return fmt.Errorf("url: %s, md5(file): %s, md5(checksum): %s error: %w",
			url, md5File, md5Checksum, errors.New("invalid md5 checksum"))
	}

	return nil
}
