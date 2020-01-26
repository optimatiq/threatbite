package browser

import (
	"regexp"
	"strings"

	aho "github.com/BobuSumisu/aho-corasick"
	"github.com/avct/uasurfer"
	"github.com/labstack/gommon/log"
)

// BrowserNames for comparison
const (
	BrowserUnknown = "Unknown"
	BrowserChrome  = "Chrome"
	BrowserIE      = "IE"
	BrowserSafari  = "Safari"
	BrowserFirefox = "Firefox"
	BrowserAndroid = "Android"
	BrowserOpera   = "Opera"

	DevicePhone = "Phone"

	OSAndroid = "Android"

	MinChromeMajorVersion  = 50
	MinIEMajorVersion      = 16
	MinFirefoxMajorVersion = 60
	MinAndroidMajorVersion = 5
	MinSafariMajorVersion  = 10
	MinOperaMajorVersion   = 10
)

// Browser information about browser name and version taken from user agent header.
// IsOld indicates that browser is considered as very old.
type Browser struct {
	Name    string
	Version struct {
		Major int
		Minor int
		Patch int
	}
	IsOld bool
}

// OS information about operating system taken from user agent header.
type OS struct {
	Name    string
	Version struct {
		Major int
		Minor int
		Patch int
	}
	Platform string
}

// UserAgent hold specific data for current uer agent.
// github.com/avct/uasurfer is used to parse data.
type UserAgent struct {
	Lowercase string
	Browser   Browser
	OS        OS
	Device    string
}

// GetUserAgent returns information about the browser/operating system and device based on user agent header value.
func GetUserAgent(agent string) *UserAgent {
	ua := uasurfer.Parse(agent)

	userAgent := &UserAgent{}
	userAgent.Lowercase = strings.ToLower(agent)
	userAgent.OS.Name = ua.OS.Name.StringTrimPrefix()
	userAgent.OS.Platform = ua.OS.Platform.StringTrimPrefix()
	userAgent.OS.Version = ua.OS.Version
	userAgent.Device = ua.DeviceType.StringTrimPrefix()
	userAgent.Browser.Name = ua.Browser.Name.StringTrimPrefix()
	userAgent.Browser.Version = ua.Browser.Version
	userAgent.Browser.IsOld = userAgent.isOldBrowser()

	return userAgent
}

// IsOldBrowser return bool value indicating if browser is outdated
func (ua *UserAgent) isOldBrowser() bool {
	switch {
	case ua.Browser.Name == BrowserIE:
		if ua.Browser.Version.Major < MinIEMajorVersion {
			return true
		}
	case ua.Browser.Name == BrowserChrome:
		if ua.Device == DevicePhone && ua.OS.Name == OSAndroid && ua.OS.Version.Major >= 8 && ua.Browser.Version.Major < 60 {
			return true
		} else if ua.Browser.Version.Major < MinChromeMajorVersion {
			return true
		}
	case ua.Browser.Name == BrowserFirefox:
		if ua.Browser.Version.Major < MinFirefoxMajorVersion {
			return true
		}
	case ua.Browser.Name == BrowserAndroid:
		if ua.Browser.Version.Major < MinAndroidMajorVersion {
			return true
		}
	case ua.Browser.Name == BrowserSafari:
		if ua.Browser.Version.Major < MinSafariMajorVersion {
			return true
		}
	case ua.Browser.Name == BrowserOpera:
		if ua.Browser.Version.Major < MinOperaMajorVersion {
			return true
		}
	}

	return false
}

var trie = aho.NewTrieBuilder().
	AddStrings([]string{
		"12soso", "192.comagent", "1noonbot", "1on1searchbot",
		"3d_search", "3de_search2", "3g bot", "3gse",
		"50.nu", "a1 sitemap generator", "a1 website download", "a6-indexer",
		"aasp", "abachobot", "abonti", "abotemailsearch",
		"aboundex", "aboutusbot", "accmonitor compliance server", "accoon",
		"achulkov.net page walker", "acme.spider", "acoonbot", "acquia-crawler",
		"activetouristbot", "ad muncher", "adamm bot", "adbeat_bot",
		"adminshop.com", "advanced email extractor", "aesop_com_spiderman", "aespider",
		"af knowledge now verity spider", "aggregator:vocus", "ah-ha.com crawler", "ahrefsbot",
		"aibot", "aidu", "aihitbot", "aipbot",
		"aisiid", "aitcsrobot/1.1", "ajsitemap", "akamai-sitesnapshot",
		"alexawebsearchplatform", "alexfdownload", "alexibot", "alkalinebot",
		"all acronyms bot", "alpha search agent", "amerla search bot", "amfibibot",
		"ampmppc.com", "amznkassocbot", "anemone", "anonymous",
		"anotherbot", "answerbot", "answerbus", "answerchase prove",
		"antbot", "antibot", "antisantyworm", "antro.net",
		"aonde-spider", "aport", "appengine-google", "appid: s~stremor-crawler-",
		"aqua_products", "arabot", "arachmo", "arachnophilia",
		"aria equalizer", "arianna.libero.it", "arikus_spider", "art-online.com",
		"artavisbot", "artera", "asaha search engine turkey", "ask",
		"aspider", "aspseek", "asterias", "astrofind",
		"athenusbot", "atlocalbot", "atomic_email_hunter", "attach",
		"attrakt", "attributor", "augurfind", "auresys",
		"autobaron crawler", "autoemailspider", "autowebdir", "avsearch-",
		"axfeedsbot", "axonize-bot", "ayna", "b2w",
		"backdoorbot", "backrub", "backstreet browser", "backweb",
		"baidu", "bandit", "batchftp", "baypup",
		"bdfetch", "becomebot", "becomejpbot", "beetlebot",
		"bender", "besserscheitern-crawl", "betabot", "big brother",
		"big data", "bigado.com", "bigcliquebot", "bigfoot",
		"biglotron", "bilbo", "bilgibetabot", "bilgibot",
		"bintellibot", "bitlybot", "bitvouseragent", "bizbot003",
		"bizbot04", "bizworks retriever", "black hole", "black.hole",
		"blackbird", "blackmask.net search engine", "blackwidow", "bladder fusion",
		"blaiz-bee", "blexbot", "blinkx", "blitzbot",
		"blog conversation project", "blogmyway", "blogpulselive", "blogrefsbot",
		"blogscope", "blogslive", "bloobybot", "blowfish",
		"blt", "bnf.fr_bot", "boaconstrictor", "boardreader",
		"boi_crawl_00", "boia-scan-agent", "boia.org", "boitho",
		"bookmark buddy bookmark checker", "bookmark search tool", "bosug", "bot apoena",
		"botalot", "botrighthere", "botswana", "bottybot",
		"bpbot", "braintime_search", "brokenlinkcheck.com", "browseremulator",
		"browsermob", "bruinbot", "bsearchr&d", "bspider",
		"btbot", "btsearch", "bubing", "buddy",
		"buibui", "buildcms crawler", "builtbottough", "bullseye",
		"bumblebee", "bunnyslippers", "buscadorclarin", "buscaplus robi",
		"butterfly", "buyhawaiibot", "buzzbot", "byindia",
		"byspider", "byteserver", "bzbot", "c r a w l 3 r",
		"cacheblaster", "caddbot", "cafi", "camcrawler",
		"camelstampede", "canon-webrecord", "careerbot", "cataguru",
		"catchbot", "cazoodle", "ccbot", "ccgcrawl",
		"ccubee", "cd-preload", "ce-preload", "cegbfeieh",
		"cerberian drtrs", "cert figleafbot", "cfetch", "cfnetwork",
		"chameleon", "charlotte", "check&get", "checkbot",
		"checklinks", "cheesebot", "chemiede-nodebot", "cherrypicker",
		"chilkat", "chinaclaw", "cipinetbot", "cis455crawler",
		"citeseerxbot", "cizilla", "clariabot", "climate ark",
		"climateark spider", "cliqzbot", "clshttp", "clushbot",
		"coast scan engine", "coast webmaster pro", "coccoc", "collapsarweb",
		"collector", "colocrossing", "combine", "connectsearch",
		"conpilot", "contentsmartz", "contextad bot", "contype",
		"cookienet", "coolbot", "coolcheck", "copernic",
		"copier", "copyrightcheck", "core-project", "cosmos",
		"covario-ids", "cowbot-", "cowdog bot", "crabbybot",
		"craftbot@yahoo.com", "crawl_application", "crawler.kpricorn.org", "crawler43.ejupiter.com",
		"crawler4j", "crawler@", "crawler_for_infomine", "crawly",
		"creativecommons", "crescent", "cs-crawler", "cse html validator",
		"cshttpclient", "cuasarbot", "culsearch", "curl",
		"custo", "cvaulev", "cyberdog", "cybernavi_webget",
		"cyberpatrol sitecat webbot", "cyberspyder", "cydralspider", "d1garabicengine",
		"datacha0s", "datafountains", "dataparksearch", "dataprovider.com",
		"datascape robot", "dataspearspiderbot", "dataspider", "dattatec.com",
		"daumoa", "dblbot", "dcpbot", "declumbot",
		"deepindex", "deepnet crawler", "deeptrawl", "dejan",
		"del.icio.us-thumbnails", "deltascan", "delvubot", "der gro§e bildersauger",
		"der große bildersauger", "deusu", "dfs-fetch", "diagem",
		"diamond", "dibot", "didaxusbot", "digext",
		"digger", "digi-rssbot", "digitalarchivesbot", "digout4u",
		"diibot", "dillo", "dir_snatch.exe", "disco",
		"distilled-reputation-monitor", "djangotraineebot", "dkimrepbot", "dmoz downloader",
		"docomo", "dof-verify", "domaincrawler", "domainscan",
		"domainwatcher bot", "dotbot", "dotspotsbot", "dow jones searchbot",
		"download", "doy", "dragonfly", "drip",
		"drone", "dtaagent", "dtsearchspider", "dumbot",
		"dwaar", "dxseeker", "e-societyrobot", "eah",
		"earth platform indexer", "earth science educator  robot", "easydl", "ebingbong",
		"ec2linkfinder", "ecairn-grabber", "ecatch", "echoosebot",
		"edisterbot", "edugovsearch", "egothor", "eidetica.com",
		"eirgrabber", "elblindo the blind bot", "elisabot", "ellerdalebot",
		"email exractor", "emailcollector", "emailleach", "emailsiphon",
		"emailwolf", "emeraldshield", "empas_robot", "enabot",
		"endeca", "enigmabot", "enswer neuro bot", "enter user-agent",
		"entitycubebot", "erocrawler", "estylesearch", "esyndicat bot",
		"eurosoft-bot", "evaal", "eventware", "everest-vulcan inc.",
		"exabot", "exactsearch", "exactseek", "exooba",
		"exploder", "express webpictures", "extractor", "eyenetie",
		"ez-robot", "ezooms", "f-bot test pilot", "factbot",
		"fairad client", "falcon", "fast data search document retriever", "fast esp",
		"fast-search-engine", "fastbot crawler", "fastbot.de crawler", "fatbot",
		"favcollector", "faviconizer", "favorites sweeper", "fdm",
		"fdse robot", "fedcontractorbot", "fembot", "fetch api request",
		"fetch_ici", "fgcrawler", "filangy", "filehound",
		"findanisp.com_isp_finder", "findlinks", "findweb", "firebat",
		"firstgov.gov search", "flaming attackbot", "flamingo_searchengine", "flashcapture",
		"flashget", "flickysearchbot", "fluffy the spider", "flunky",
		"focused_crawler", "followsite", "foobot", "fooooo_web_video_crawl",
		"fopper", "formulafinderbot", "forschungsportal", "fr_crawler",
		"francis", "freewebmonitoring sitechecker", "freshcrawler", "freshdownload",
		"freshlinks.exe", "friendfeedbot", "frodo.at", "froggle",
		"frontpage", "froola bot", "fu-nbi", "full_breadth_crawler",
		"funnelback", "furlbot", "g10-bot", "gaisbot",
		"galaxybot", "gazz", "gbplugin", "generate_infomine_category_classifiers",
		"genevabot", "geniebot", "genieo", "geomaxenginebot",
		"geometabot", "geonabot", "geovisu", "germcrawler",
		"gethtmlcontents", "getleft", "getright", "getsmart",
		"geturl.rexx", "getweb!", "giant", "gigablastopensource",
		"gigabot", "girafabot", "gleamebot", "gnome-vfs",
		"go!zilla", "go-ahead-got-it", "go-http-client", "goforit.com",
		"goforitbot", "gold crawler", "goldfire server", "golem",
		"goodjelly", "gordon-college-google-mini", "goroam", "goseebot",
		"gotit", "govbot", "gpu p2p crawler", "grabber",
		"grabnet", "grafula", "grapefx", "grapeshot",
		"grbot", "greenyogi", "gromit", "grub",
		"gsa", "gslfbot", "gulliver", "gulperbot",
		"gurujibot", "gvc business crawler", "gvc crawler", "gvc search bot",
		"gvc web crawler", "gvc weblink crawler", "gvc world links", "gvcbot.com",
		"happyfunbot", "harvest", "hatena antenna", "hawler",
		"hcat", "hclsreport-crawler", "hd nutch agent", "header_test_client",
		"healia", "helix", "here will be link to crawler site", "heritrix",
		"hiscan", "hisoftware accmonitor server", "hisoftware accverify", "hitcrawler",
		"hivabot", "hloader", "hmsebot", "hmview",
		"hoge", "holmes", "homepagesearch", "hooblybot-image",
		"hoowwwer", "hostcrawler", "hsft - link scanner", "hsft - lvu scanner",
		"hslide", "ht://check", "htdig", "html link validator",
		"htmlparser", "httplib", "httrack", "huaweisymantecspider",
		"hul-wax", "humanlinks", "hyperestraier", "hyperix",
		"ia_archiver", "iaarchiver-", "ibuena", "icab",
		"icds-ingestion", "ichiro", "icopyright conductor", "ieautodiscovery",
		"iecheck", "ihwebchecker", "iiitbot", "iim_405",
		"ilsebot", "iltrovatore", "image stripper", "image sucker",
		"image-fetcher", "imagebot", "imagefortress", "imageshereimagesthereimageseverywhere",
		"imagevisu", "imds_monitor", "imo-google-robot-intelink", "inagist.com url crawler",
		"indexer", "industry cortex webcrawler", "indy library", "indylabs_marius",
		"inelabot", "inet32 ctrl", "inetbot", "info seeker",
		"infolink", "infomine", "infonavirobot", "informant",
		"infoseek sidewinder", "infotekies", "infousabot", "ingrid",
		"inktomi", "insightscollector", "insightsworksbot", "inspirebot",
		"insumascout", "intelix", "intelliseek", "interget",
		"internet ninja", "internet radio crawler", "internetlinkagent", "interseek",
		"ioi", "ip-web-crawler.com", "ipadd bot", "ips-agent",
		"ipselonbot", "iria", "irlbot", "iron33",
		"isara", "isearch", "isilox", "istellabot",
		"its-learning crawler", "iu_csci_b659_class_crawler", "ivia", "jadynave",
		"java", "jbot", "jemmathetourist", "jennybot",
		"jetbot", "jetbrains omea pro", "jetcar", "jim",
		"jobo", "jobspider_ba", "joc", "joedog",
		"joyscapebot", "jspyda", "junut bot", "justview",
		"jyxobot", "k.s.bot", "kakclebot", "kalooga",
		"katatudo-spider", "kbeta1", "keepni web site monitor", "kenjin.spider",
		"keybot translation-search-machine", "keywenbot", "keyword density", "keyword.density",
		"kinjabot", "kitenga-crawler-bot", "kiwistatus", "kmbot-",
		"kmccrew bot search", "knight", "knowitall", "knowledge engine",
		"knowledge.com", "koepabot", "koninklijke", "korniki",
		"krowler", "ksbot", "kuloko-bot", "kulturarw3",
		"kummhttp", "kurzor", "kyluka crawl", "l.webis",
		"labhoo", "labourunions411", "lachesis", "lament",
		"lamerexterminator", "lapozzbot", "larbin", "lbot",
		"leaptag", "leechftp", "leechget", "letscrawl.com",
		"lexibot", "lexxebot", "lftp", "libcrawl",
		"libiviacore", "libw", "likse", "linguee bot",
		"link checker", "link validator", "link_checker", "linkalarm",
		"linkbot", "linkcheck by siteimprove.com", "linkcheck scanner", "linkchecker",
		"linkdex.com", "linkextractorpro", "linklint", "linklooker",
		"linkman", "links sql", "linkscan", "linksmanager.com_bot",
		"linksweeper", "linkwalker", "litefinder", "litlrbot",
		"little grabber at skanktale.com", "livelapbot", "lm harvester", "lmqueuebot",
		"lnspiderguy", "loadtimebot", "localcombot", "locust",
		"lolongbot", "lookbot", "lsearch", "lssbot",
		"lt scotland checklink", "ltx71.com", "lwp", "lycos_spider",
		"lydia entity spider", "lynnbot", "lytranslate", "mag-net",
		"magnet", "magpie-crawler", "magus bot", "mail.ru",
		"mainseek_bot", "mammoth", "map robot", "markwatch",
		"masagool", "masidani_bot_", "mass downloader", "mata hari",
		"mata.hari", "matentzn at cs dot man dot ac dot uk", "maxamine.com--robot", "maxamine.com-robot",
		"maxomobot", "mcbot", "medrabbit", "megite",
		"memacbot", "memo", "mendeleybot", "mercator-",
		"mercuryboard_user_agent_sql_injection.nasl", "metacarta", "metaeuro web search", "metager2",
		"metagloss", "metal crawler", "metaquerier", "metaspider",
		"metaspinner", "metauri", "mfcrawler", "mfhttpscan",
		"midown tool", "miixpc", "mini-robot", "minibot",
		"minirank", "mirror", "missigua locator", "mister pix",
		"mister.pix", "miva", "mj12bot", "mnogosearch",
		"mod_accessibility", "moduna.com", "moget", "mojeekbot",
		"monkeycrawl", "moses", "mowserbot", "mqbot",
		"mse360", "msindianwebcrawl", "msmobot", "msnptc",
		"msrbot", "mt-soft", "multitext", "my-heritrix-crawler",
		"my_little_searchengine_project", "myapp", "mycompanybot", "mycrawler",
		"myengines-us-bot", "myfamilybot", "myra", "nabot",
		"najdi.si", "nambu", "nameprotect", "nasa search",
		"natchcvs", "natweb-bad-link-mailer", "naver", "navroad",
		"nearsite", "nec-meshexplorer", "neosciocrawler", "nerdbynature.bot",
		"nerdybot", "nerima-crawl-", "nessus", "nestreader",
		"net vampire", "net::trackback", "netants", "netcarta cyberpilot pro",
		"netcraft", "netexperts", "netid.com bot", "netmechanic",
		"netprospector", "netresearchserver", "netseer", "netshift=",
		"netsongbot", "netsparker", "netspider", "netsrcherp",
		"netzip", "newmedhunt", "news bot", "news_search_app",
		"newsgatherer", "newsgroupreporter", "newstrovebot", "nextgensearchbot",
		"nextthing.org", "nicebot", "nicerspro", "niki-bot",
		"nimblecrawler", "nimbus-1", "ninetowns", "ninja",
		"njuicebot", "nlese", "nogate", "norbert the spider",
		"noteworthybot", "npbot", "nrcan intranet crawler", "nsdl_search_bot",
		"nu_tch", "nuggetize.com bot", "nusearch spider", "nutch",
		"nwspider", "nymesis", "nys-crawler", "objectssearch",
		"obot", "obvius external linkcheck", "ocelli", "octopus",
		"odp entries t_st", "oegp", "offline navigator", "offline.explorer",
		"ogspider", "omiexplorer_bot", "omniexplorer", "omnifind",
		"omniweb", "onetszukaj", "online link validator", "oozbot",
		"openbot", "openfind", "openintelligencedata", "openisearch",
		"openlink virtuoso rdf crawler", "opensearchserver_bot", "opidig", "optidiscover",
		"oracle secure enterprise search", "oracle ultra search", "orangebot", "orisbot",
		"ornl_crawler", "ornl_mercury", "osis-project.jp", "oso",
		"outfoxbot", "outfoxmelonbot", "owler-bot", "owsbot",
		"ozelot", "p3p client", "page_verifier", "pagebiteshyperbot",
		"pagebull", "pagedown", "pagefetcher", "pagegrabber",
		"pagerank monitor", "pamsnbot.htm", "panopy bot", "panscient.com",
		"pansophica", "papa foto", "paperlibot", "parasite",
		"parsijoo", "pathtraq", "pattern", "patwebbot",
		"pavuk", "paxleframework", "pbbot", "pcbrowser",
		"pcore-http", "pd-crawler", "penthesila", "perform_crawl",
		"perman", "personal ultimate crawler", "php version tracker", "phpcrawl",
		"phpdig", "picosearch", "pieno robot", "pipbot",
		"pipeliner", "pita", "pixfinder", "piyushbot",
		"planetwork bot search", "plucker", "plukkie", "plumtree",
		"pockey", "pocohttp", "pogodak.ba", "pogodak.co.yu",
		"poirot", "polybot", "pompos", "poodle predictor",
		"popscreenbot", "postpost", "privacyfinder", "projectwf-java-test-crawler",
		"propowerbot", "prowebwalker", "proxem websearch", "proximic",
		"proxy crawler", "psbot", "pss-bot", "psycheclone",
		"pub-crawler", "pucl", "pulsebot", "pump",
		"pwebot", "python", "qeavis agent", "qfkbot",
		"qualidade", "qualidator.com bot", "quepasacreep", "queryn metasearch",
		"queryn.metasearch", "quest.durato", "quintura-crw", "qunarbot",
		"qwantify", "qweery_robot.txt_checkbot", "qweerybot", "r2ibot",
		"r6_commentreader", "r6_feedfetcher", "r6_votereader", "rabot",
		"radian6", "radiation retriever", "rampybot", "rankivabot",
		"rankur", "rational sitecheck", "rcstartbot", "realdownload",
		"reaper", "rebi-shoveler", "recorder", "redbot",
		"redcarpet", "reget", "repomonkey", "research robot",
		"riddler", "riight", "risenetbot", "riverglassscanner",
		"robopal", "robosourcer", "robotek", "robozilla",
		"roger", "rome client", "rondello", "rotondo",
		"roverbot", "rpt-httpclient", "rtgibot", "rufusbot",
		"runnk online rss reader", "runnk rss aggregator", "s2bot", "safaribookmarkchecker",
		"safednsbot", "safetynet robot", "saladspoon", "sapienti",
		"sapphireweb", "sbider", "sbl-bot", "scfcrawler",
		"scich", "scientificcommons.org", "scollspider", "scooperbot",
		"scooter", "scoutjet", "scrapebox", "scrapy",
		"scrawltest", "scrubby", "scspider", "scumbot",
		"search publisher", "search x-bot", "search-channel", "search-engine-studio",
		"search.kumkie.com", "search.updated.com", "search.usgs.gov", "searcharoo.net",
		"searchblox", "searchbot", "searchengine", "searchhippo.com",
		"searchit-bot", "searchmarking", "searchmarks", "searchmee!",
		"searchmee_v", "searchmining", "searchnowbot", "searchpreview",
		"searchspider.com", "searqubot", "seb spider", "seekbot",
		"seeker.lookseek.com", "seeqbot", "seeqpod-vertical-crawler", "selflinkchecker",
		"semager", "semanticdiscovery", "semantifire", "semisearch",
		"semrushbot", "seoengworldbot", "seokicks", "seznambot",
		"shablastbot", "shadowwebanalyzer", "shareaza", "shelob",
		"sherlock", "shim-crawler", "shopsalad", "shopwiki",
		"showlinks", "showyoubot", "siclab", "silk",
		"simplepie", "siphon", "sitebot", "sitecheck",
		"sitefinder", "siteguardbot", "siteorbiter", "sitesnagger",
		"sitesucker", "sitesweeper", "sitexpert", "skimbot",
		"skimwordsbot", "skreemrbot", "skywalker", "sleipnir",
		"slow-crawler", "slysearch", "smart-crawler", "smartdownload",
		"smarte bot", "smartwit.com", "snake", "snap.com beta crawler",
		"snapbot", "snappreviewbot", "snappy", "snookit",
		"snooper", "snoopy", "societyrobot", "socscibot",
		"soft411 directory", "sogou", "sohu agent", "sohu-search",
		"sokitomi crawl", "solbot", "sondeur", "sootle",
		"sosospider", "space bison", "space fung", "spacebison",
		"spankbot", "spanner", "spatineo monitor controller", "spatineo serval controller",
		"spatineo serval getmapbot", "special_archiver", "speedy", "sphere scout",
		"sphider", "spider.terranautic.net", "spiderengine", "spiderku",
		"spiderman", "spinn3r", "spinne", "sportcrew-bot",
		"sproose", "spyder3.microsys.com", "sq webscanner", "sqlmap",
		"squid-prefetch", "squidclamav_redirector", "sqworm", "srevbot",
		"sslbot", "ssm agent", "stackrambler", "stardownloader",
		"statbot", "statcrawler", "statedept-crawler", "steeler",
		"stegmann-bot", "stero", "stripper", "stumbler",
		"suchclip", "sucker", "sumeetbot", "sumitbot",
		"summizebot", "summizefeedreader", "sunrise xp", "superbot",
		"superhttp", "superlumin downloader", "superpagesbot", "supremesearch.net",
		"supybot", "surdotlybot", "surf", "surveybot",
		"suzuran", "swebot", "swish-e", "sygolbot",
		"synapticwalker", "syntryx ant scout chassis pheromone", "systemsearch-robot", "szukacz",
		"s~stremor-crawler", "t-h-u-n-d-e-r-s-t-o-n-e", "tailrank", "takeout",
		"talkro web-shot", "tamu_crawler", "tapuzbot", "tarantula",
		"targetblaster.com", "targetyournews.com bot", "tausdatabot", "taxinomiabot",
		"teamsoft wininet component", "tecomi bot", "teezirbot", "teleport",
		"telesoft", "teradex mapper", "teragram_crawler", "terrawizbot",
		"testbot", "testing of bot", "textbot", "thatrobotsite.com",
		"the dyslexalizer", "the intraformant", "the.intraformant", "thenomad",
		"theophrastus", "theusefulbot", "thumbbot", "thumbnail.cz robot",
		"thumbshots-de-bot", "tigerbot", "tighttwatbot", "tineye",
		"titan", "to-dress_ru_bot_", "to-night-bot", "tocrawl",
		"topicalizer", "topicblogs", "toplistbot", "topserver php",
		"topyx-crawler", "touche", "tourlentascanner", "tpsystem",
		"traazi", "transgenikbot", "travel-search", "travelbot",
		"travellazerbot", "treezy", "trendiction", "trex",
		"tridentspider", "trovator", "true_robot", "tscholarsbot",
		"tsm translation-search-machine", "tswebbot", "tulipchain", "turingos",
		"turnitinbot", "tutorgigbot", "tweetedtimes bot", "tweetmemebot",
		"tweezler", "twengabot", "twice", "twikle",
		"twinuffbot", "twisted pagegetter", "twitturls", "twitturly",
		"tygobot", "tygoprowler", "typhoeus", "u.s. government printing office",
		"uberbot", "ucb-nutch", "udmsearch", "ufam-crawler-",
		"ultraseek", "unchaos", "unidentified", "unisterbot",
		"unitek uniengine", "universalsearch", "unwindfetchor", "uoftdb_experiment",
		"updated", "uptimebot/1.0", "url control", "url-checker",
		"url_gather", "urlappendbot", "urlblaze", "urlchecker",
		"urlck", "urldispatcher", "urlspiderpro", "urly warning",
		"urly.warning", "usaf afkn k2spider", "usasearch", "uss-cosmix",
		"usyd-nlp-spider", "vacobot", "vacuum", "vadixbot",
		"vagabondo", "valkyrie", "vbseo", "vci webviewer vci webviewer win32",
		"verbstarbot", "vericitecrawler", "verifactrola", "verity-url-gateway",
		"vermut", "versus crawler", "versus.integis.ch", "viasarchivinginformation.html",
		"vipr", "virus-detector", "virus_detector", "visbot",
		"vishal for clia", "visweb", "vital search'n urchin", "vlad",
		"vlsearch", "vmbot", "vocusbot", "voideye",
		"voil", "voilabot", "vortex", "voyager",
		"vspider", "vuhuvbot/1.0", "w3c-webcon", "w3c_unicorn",
		"w3search", "wacbot", "wanadoo", "wastrix",
		"water conserve portal", "water conserve spider", "watzbot", "wauuu",
		"wavefire", "waypath", "wazzup", "wbdbot",
		"wbsrch", "web ceo online robot", "web crawler", "web downloader",
		"web image collector", "web link validator", "web magnet", "web site downloader",
		"web sucker", "web-agent", "web-sniffer", "web.image.collector",
		"webaltbot", "webauto", "webbot", "webbul-bot",
		"webcapture", "webcheck", "webclipping.com", "webcollage",
		"webcopier", "webcopy", "webcorp", "webcrawl.net",
		"webcrawler", "webdatacentrebot", "webdownloader for x", "webdup",
		"webemailextrac", "webenhancer", "webfetch", "webgather",
		"webgo is", "webgobbler", "webimages", "webinator-search2",
		"webinator-wbi", "webindex", "weblayers", "webleacher",
		"weblexbot", "weblinker", "weblyzard", "webmastercoffee",
		"webmasterworld extractor", "webmasterworldforumbot", "webminer", "webmoose",
		"webot", "webpix", "webreaper", "webripper",
		"websauger", "webscan", "websearchbench", "website",
		"webspear", "websphinx", "webspider", "webster",
		"webstripper", "webtrafficexpress", "webtrends link analyzer", "webvac",
		"webwalk", "webwasher", "webwatch", "webwhacker",
		"webxm", "webzip", "weddings.info", "wenbin",
		"wep search", "wepa", "werelatebot", "wget",
		"whacker", "whirlpool web engine", "whowhere robot", "widow",
		"wikiabot", "wikio", "wikiwix-bot-", "winhttp",
		"wire", "wisebot", "wisenutbot", "wish-la",
		"wish-project", "wisponbot", "wmcai-robot", "wminer",
		"wmsbot", "woriobot", "worldshop", "worqmada",
		"wotbox", "wume_crawler", "www collector", "www-collector-e",
		"www-mechanize", "wwwoffle", "wwwrobot", "wwwster",
		"wwwwanderer", "wwwxref", "wysigot", "x-clawler",
		"x-crawler", "xaldon", "xenu", "xerka metabot",
		"xerka webbot", "xget", "xirq", "xmarksfetch",
		"xqrobot", "y!j", "yacy.net", "yacybot",
		"yanga worldsearch bot", "yarienavoir.net", "yasaklibot", "yats crawler",
		"ybot", "yebolbot", "yellowjacket", "yeti",
		"yolinkbot", "yooglifetchagent", "yoono", "yottacars_bot",
		"yourls", "z-add link checker", "zagrebin", "zao",
		"zedzo.validate", "zermelo", "zeus", "zibber-v",
		"zimeno", "zing-bottabot", "zipppbot", "zongbot",
		"zoomspider", "zotag search", "zsebot", "zuibot",
	}).Build()

// IsBotUserAgent return if UserAgent is a bot
func IsBotUserAgent(agent string) bool {
	matches := trie.MatchString(strings.ToLower(agent))
	if len(matches) > 0 {
		log.Debugf("[IsBotUserAgent] ip: %s", matches)
		return true
	}
	return false
}

var reMobileUserAgent = regexp.MustCompile("(?:hpw|i|web)os|alamofire|alcatel|amoi|android|avantgo|blackberry|blazer|cell|cfnetwork|darwin|dolfin|dolphin|fennec|htc|ip(?:hone|od|ad)|ipaq|j2me|kindle|midp|minimo|mobi|motorola|nec-|netfront|nokia|opera m(ob|in)i|palm|phone|pocket|portable|psp|silk-accelerated|skyfire|sony|ucbrowser|up.browser|up.link|windows ce|xda|zte|zune")

// IsMobileUserAgent Checks if User-Agent is from mobile device
func IsMobileUserAgent(agent string) bool {
	if reMobileUserAgent.MatchString(strings.ToLower(agent)) {
		log.Debugf("[IsMobileUserAgent] ip: %s")
		return true
	}
	return false
}

var reScriptUserAgent = regexp.MustCompile("curl|wget|collectd|python|urllib|java|jakarta|httpclient|phpcrawl|libwww|perl|go-http|okhttp|lua-resty|winhttp|awesomium")

// IsScriptUserAgent check if UserAgent comes from script
func IsScriptUserAgent(agent string) bool {
	if reScriptUserAgent.MatchString(strings.ToLower(agent)) {
		log.Debugf("[IsScriptUserAgent] ip: %s")
		return true
	}
	return false
}
