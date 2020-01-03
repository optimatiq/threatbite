package ip

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	aho "github.com/BobuSumisu/aho-corasick"
	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/ip/datasource"
)

type datacenter struct {
	ipnet *datasource.IPNet
	geoip geoip
}

func newDC(geoip geoip, source datasource.DataSource) *datacenter {
	list := datasource.NewIPNet(source, "datacenter")
	return &datacenter{
		ipnet: list,
		geoip: geoip,
	}
}

var reIsDC = regexp.MustCompile("server|vps|cloud|web|hosting|virt")

var trie = aho.NewTrieBuilder().
	AddStrings([]string{
		"1&1", "1and1", "1gb", "21vianet", "23media", "23vnet", "2dayhost", "3nt",
		"4rweb", "abdicar", "abelohost", "acceleratebiz", "accelerated", "acenet", "acens", "activewebs.dk",
		"adhost", "advancedhosters", "advania", "ainet", "airnet group", "akamai", "alfahosting", "alibaba",
		"allhostshop", "almahost", "alog", "alpharacks", "altushost", "alvotech", "amanah", "amazon",
		"amerinoc", "anexia", "apollon", "applied", "ardis", "ares", "argeweb", "argon",
		"aruba", "arvixe", "atman", "atomohost", "availo.no", "avantehosting", "avguro", "awknet",
		"aws", "azar-a", "azure", "b2 net", "basefarm", "beget", "best-hosting", "beyond",
		"bhost", "biznes-host", "blackmesh", "blazingfast", "blix", "blue mile", "blueconnex", "bluehost",
		"bodhost", "braslink", "brinkster", "budgetbytes", "burstnet", "business", "buyvm", "calpop",
		"canaca-com", "carat networks", "cari", "ccpg", "ceu", "ch-center", "choopa", "cinipac",
		"cirrus", "cloud", "cloudflare", "cloudsigma", "cloudzilla", "co-location", "codero", "colo",
		"colo4dallas", "colo@", "colocall", "colocation", "combell", "comfoplace", "comvive", "conetix",
		"confluence", "connectingbytes", "connectingbytse", "connectria", "contabo", "continuum", "coolvds", "corponetsa",
		"creanova", "crosspointcolo", "ctrls", "cybernetic-servers", "cyberverse", "cyberwurx", "cyquator", "d-hosting",
		"data 102", "data centers", "data foundry", "data shack", "data xata", "data-centr", "data-xata", "database",
		"datacenter", "datacenterscanada", "datacheap", "dataclub", "datahata.by", "datahouse.nl", "datapipe", "datapoint",
		"datasfera", "datotel", "dedi", "dedibox", "dediserv", "dedizull", "delta bulgaria", "delta-x",
		"deltahost", "demos", "deninet hungary", "depo40", "depot", "deziweb", "dfw-datacenter", "dhap",
		"digicube", "digital", "digitalocean", "digitalone", "digiweb", "dimenoc", "dinahosting", "directspace",
		"directvps", "dominios", "dotster", "dreamhost", "duocast", "duomenu centras lithuania", "e-commercepark", "e24cloud",
		"earthlink", "easyhost.be", "easyhost.hk", "easyname", "easyspeedy", "eboundhost", "ecatel", "ecritel",
		"edgewebhosting", "edis", "egihosting", "ehostidc", "ehostingusa", "ekvia", "elserver", "elvsoft",
		"enzu", "epiohost", "erix-colo", "esc", "esds", "eserver", "esited", "estoxy",
		"estroweb", "ethnohosting", "ethr", "eukhost", "euro-web", "eurobyte", "eurohoster", "eurovps",
		"everhost", "evovps", "fasthosts", "fastly", "fastmetrics", "fdcservers", "fiberhub", "fibermax",
		"finaltek", "firehost", "first colo", "firstvds", "flexwebhosting", "flokinet", "flops", "forpsi",
		"forta trust", "fortress", "fsdata", "galahost", "gandi", "gbps", "gearhost", "genesys",
		"giga-hosting", "gigahost", "gigenet", "glesys", "go4cloud", "godaddy", "gogrid", "goodnet",
		"google", "gorack", "gorilla", "gplhost", "grid", "gyron", "h1 host", "h1host",
		"h4hosting", "h88", "heart", "hellovps", "hetzner", "hispaweb", "hitme", "hivelocity",
		"home.pl", "homecloud", "homenet", "hopone", "hosixy", "host", "host department", "host virtual",
		"host-it", "host1plus", "hosta rica", "hostbasket", "hosteam.pl", "hosted", "hoster", "hosteur",
		"hostex", "hostex.lt", "hostgrad", "hosthane", "hostinet", "hosting", "hostinger", "hostkey",
		"hostmysite", "hostnet.nl", "hostnoc", "hostpro", "hostrevenda", "hostrocket", "hostventures", "hostway",
		"hostwinds", "hqhost", "hugeserver", "hurricane", "hyperhosting", "i3d", "iaas", "icn.bg",
		"ideal-solution.org", "idealhosting", "ihc", "ihnetworks", "ikoula", "iliad", "immedion", "imperanet",
		"inasset", "incero", "incubatec gmbh - srl", "indiana", "inferno", "infinitetech", "infinitie", "infinys",
		"infium-1", "infiumhost", "infobox", "infra", "inline", "inmotion hosting", "integrity", "interhost",
		"interracks", "interserver", "iomart", "iomart hosting ltd", "ionity", "ip exchange", "ip server", "ip serverone",
		"ipglobe", "iphouse", "ipserver", "ipx", "iqhost", "isppro", "ispserver", "ispsystem",
		"itl", "iweb", "iws", "ix-host", "ixam-hosting", "jumpline inc", "justhost", "keyweb",
		"kievhosting", "kinx", "knownsrv", "kualo", "kylos.pl", "latisys", "layered", "leaderhost",
		"leaseweb", "lightedge", "limelight", "limestone", "link11", "linode", "lionlink", "lippunerhosting",
		"liquid", "local", "locaweb", "logicworks", "loopbyte", "loopia", "loose", "lunar",
		"main-hosting", "marosnet", "masterhost", "mchost", "media temple", "melbicom", "memset", "mesh",
		"mgnhost", "micfo", "micron21", "microsoft", "midphase", "mirahost", "mirohost", "mnogobyte",
		"mojohost", "mrhost", "mrhost.biz", "multacom", "mxhost", "my247webhosting", "myh2oservers", "myhost",
		"myloc", "nano", "natro", "nbiserv", "ndchost", "nedzone.nl", "neospire", "net4",
		"netangels", "netbenefit", "netcup", "netelligent", "netgroup", "netinternet", "netio", "netirons",
		"netnation", "netplus", "netriplex", "netrouting", "netsys", "netzozeker.nl", "nforce", "nimbushosting",
		"nine.ch", "niobeweb", "nthost", "ntx", "nufuture", "nwt idc", "o2switch", "offshore",
		"one", "online", "openhosting", "optimate", "ovh", "ozhosting", "packet", "pair networks",
		"panamaserver", "patrikweb", "pce-net", "peak", "peak10", "peer 1", "perfect ip", "peron",
		"persona host", "pghosting", "planethoster", "plusserver", "plutex", "portlane", "premia", "prioritycolo",
		"private layer", "privatesystems", "profihost", "prohoster", "prolocation", "prometey", "providerdienste", "prq.se",
		"psychz", "quadranet", "quasar", "quickweb.nz", "qweb", "qwk", "r01", "rack",
		"rackforce", "rackmarkt", "rackplace", "rackspace", "racksrv", "rackvibe", "radore", "rapidhost",
		"rapidspeeds", "razor", "readyspace", "realcomm", "rebel", "redehost.br", "redstation", "reflected",
		"reg", "register", "regtons", "reliable", "rent", "rijndata.nl", "rimu", "risingnet",
		"root", "roya", "rtcomm", "ru-center", "s.r.o.", "saas", "sadecehosting", "safe",
		"sakura", "securewebs", "seeweb.it", "seflow", "selectel", "servage", "servenet", "server",
		"serverbeach", "serverboost.nl", "servercentral", "servercentric", "serverclub", "serverius", "servermania", "serveroffer",
		"serverpronto", "servers", "serverspace", "servhost", "servihosting", "servint", "servinus", "servisweb",
		"sevenl", "sharktech", "silicon valley", "simcentric", "simplecloud", "simpliq", "singlehop", "siteserver",
		"slask", "small orange", "smart-hosting", "smartape", "snel", "softlayer", "solido", "sologigabit",
		"spaceweb", "sparkstation", "sprocket", "staminus", "star-hosting", "steadfast", "steep host", "store",
		"strato", "sunnyvision", "superdata", "superhost.pl", "superhosting.bg", "supernetwork", "supreme", "swiftway",
		"switch", "switch media", "szervernet", "t-n media", "tagadab", "tailor made", "take 2", "tangram",
		"techie media", "technologies", "telecity", "tencent cloud", "tentacle", "teuno", "the bunker", "the endurance",
		"theplanet", "thorn", "thrust vps", "tierpoint", "tilaa", "titan internet", "totalin", "trabia",
		"tranquil", "transip", "travailsystems", "triple8", "trueserver.nl", "turkiye", "tuxis.nl", "twooit",
		"uadomen", "ubiquity", "uk2", "uk2group", "ukwebhosting.ltd.uk", "unbelievable", "unitedcolo", "unithost",
		"upcloud", "usonyx", "vautron", "vds64", "veesp", "velia", "velocity", "ventu",
		"versaweb", "vexxhost", "vhoster", "virpus", "virtacore", "vnet", "volia", "vooservers",
		"voxel", "voxility", "vpls", "vps", "vps4less", "vpscheap", "vpsnet", "vshosting.cz",
		"vstoike russia", "web werks", "web2objects", "webair", "webalta", "webaxys", "webcontrol", "webexxpurts",
		"webfusion", "webhoster", "webhosting", "webnx", "websitewelcome", "websupport", "webvisions", "wedos",
		"weebly", "wehostall", "wehostwebsites", "westhost", "wholesale", "wiredtree", "worldstream", "wow",
		"x10hosting", "xentime", "xiolink", "xirra gmbh", "xlhost", "xmission", "xserver", "xservers",
		"xt global", "xtraordinary", "yandex", "yeshost", "yisp", "yourcolo", "yourserver", "zare",
		"zenlayer", "zet", "zomro", "zservers",
	}).
	Build()

func (p *datacenter) isDC(ip net.IP) (bool, error) {
	isDC, err := p.ipnet.Check(ip)
	if err != nil {
		return false, fmt.Errorf("cannot run Check on %s, error: %w", ip, err)
	}
	if isDC {
		log.Debugf("[isDC] ip: %s dc: %t", ip, isDC)
		return true, nil
	}

	organisation, err := p.geoip.getCompany(ip)
	if err != nil {
		return false, fmt.Errorf("cannot run getCompanyName on %s, error: %w", ip, err)
	}
	if organisation != "" {
		matches := trie.MatchString(strings.ToLower(organisation))
		if len(matches) > 0 {
			return true, nil
		}
	}

	hostnames, err := lookupAddrWithTimeout(ip.String(), 500*time.Millisecond)
	if err != nil {
		// errors like "no such host" are normal, we don't need to pollute error logs
		log.Debugf("[isDC] ip: %s error: %s", ip, err)
		return false, nil
	}

	if reIsDC.MatchString(hostnames[0]) {
		log.Debugf("[isDC] ip: %s DC match", ip)
		return true, nil
	}

	return false, nil
}
