package browser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type UseragentSuite struct {
	suite.Suite
}

func (suite *UseragentSuite) TestParseUseragent() {
	tests := []struct {
		ua   string
		want *UserAgent
	}{
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36", &UserAgent{
			Lowercase: "mozilla/5.0 (macintosh; intel mac os x 10_14_4) applewebkit/537.36 (khtml, like gecko) chrome/73.0.3683.86 safari/537.36",
			Device:    "Computer",
			OS: OS{
				Name: "MacOSX", Platform: "Mac", Version: struct {
					Major int
					Minor int
					Patch int
				}{Major: 10, Minor: 14, Patch: 4}},
			Browser: Browser{
				Name: BrowserChrome,
				Version: struct {
					Major int
					Minor int
					Patch int
				}{Major: 73, Minor: 0, Patch: 3683}},
		}},
		{"invalid", &UserAgent{
			Lowercase: "invalid",
			Device:    "Unknown",
			OS: OS{
				Name: "Unknown", Platform: "Unknown"},
			Browser: Browser{
				Name: BrowserUnknown,
			},
		}},
	}

	for _, t := range tests {
		ua := GetUserAgent(t.ua)
		suite.Equal(t.want, ua)
	}
}

func (suite *UseragentSuite) TestParseUseragentOldBrowser() {
	tests := []struct {
		ua    string
		isOld bool
	}{
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36", false},
		{"Mozilla/5.0 (Linux; Android 4.4.2; SAMSUNG-SGH-I337 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/33.0.1750.170 Mobile Safari/537.36", true},
		{"Mozilla/5.0 (Linux; U; Android 4.0.3; de-ch; HTC Sensation Build/IML74K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30", true},
		{"Mozilla/5.0 (Linux; Android 6.0; ALE-L21) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.80 Mobile Safari/537.36", false},
		{"Mozilla/5.0 (Linux; Android 6.0; XT1072 Build/MPBS24.65-34-5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Mobile Safari/537.36", false},
		{"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; MDDRJS; rv:11.0) like Gecko", true},
		{"Mozilla/5.0 (Linux; Android 5.0.2; SAMSUNG SM-T535 Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/9.2 Chrome/67.0.3396.87 Safari/537.36", false},
		{"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36", true},
		{"Mozilla/5.0 (iPad; CPU OS 5_1 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9B176 Safari/7534.48.3", true},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36 OPR/56.0.3051.116", false},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 7_0_2 like Mac OS X) AppleWebKit/537.51.1 (KHTML, like Gecko) Version/7.0 Mobile/11A501 Safari/9537.53", true},
		{"Mozilla/5.0 (Windows Phone 10.0; Android 6.0.1; NOKIA; Lumia 830) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Mobile Safari/537.36 Edge/14.14393", true},
		{"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 YaBrowser/19.3.1.828 Yowser/2.5 Safari/537.36", false},
		{"Mozilla/5.0 (Windows NT 6.3; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0", false},
		{"Mozilla/5.0 (Android 7.0; Mobile; rv:58.0) Gecko/58.0 Firefox/58.0", true},
		{"Opera/9.80 (J2ME/iPhone;Opera Mini/5.0.019802/886; U; ja)Presto/2.4.15", true},
		{"Opera/9.80 (J2ME/iPhone;Opera Mini/5.0.019802/886; U; ja)Presto/2.4.15", true}, // duplicate to check if once set properly
		{"Mozilla/5.0 (Linux; Android 8.0.0; SM-G930F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.3729.136 Mobile Safari/537.36", true},
		{"Mozilla/5.0 (Linux; Android 7.0.0; SM-G930F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.3729.136 Mobile Safari/537.36", true},
		{"Not a browser UA", false},
		{"", false},
		// should be true - fix uasurfer to support such browsers (now returns as Unknown)
		{"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.17) Gecko/20080919 K-Meleon/1.5.1", false},
		{"Mozilla/5.0 (Windows; U; Windows NT 5.1; pt-BR) AppleWebKit/533.3 (KHTML, like Gecko) QtWeb Internet Browser/3.7 http://www.QtWeb.net", false},
		{"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.2a1pre) Gecko/20090316 Minefield/3.2a1pre", false},
	}

	for _, t := range tests {
		ua := GetUserAgent(t.ua)
		suite.Equal(t.isOld, ua.Browser.IsOld, t)
	}
}

func TestUseragentSuite(t *testing.T) {
	suite.Run(t, new(UseragentSuite))
}
