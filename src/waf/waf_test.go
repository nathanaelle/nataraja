package waf


import	(
	"testing"
	"strings"
)




var UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:40.0) Gecko/20100101 Firefox/40.0"


var bad_robots = []string {
	"webmole",
	"wisenutbot",
	"prowebwalker",
	"hanzoweb",
	"email",
	"toata dragostea mea pentru diavola",
	"gameBoy, powered by nintendo",
	"missigua",
	"poe-component-client",
	"emailsiphon",
	"adsarobot",
	"under the rainbow 2.",
	"nessus",
	"floodgate",
	"email extractor",
	"webaltbot",
	"contactbot/",
	"butch__2.1.1",
	"pe 1.4",
	"indy library",
	"autoemailspider",
	"mozilla/3.mozilla/2.01",
	"fantombrowser",
	"digout4uagent",
	"panscient.com",
	"telesoft",
	"; widows",
	"converacrawler",
	"www.weblogs.com",
	"murzillo compatible",
	"isc systems irc search 2.1",
	"emailmagnet",
	"microsoft url control",
	"datacha0s",
	"emailwolf",
	"production bot",
	"sitesnagger",
	"webbandit",
	"web by mail",
	"faxobot",
	"grub crawler",
	"jakarta",
	"eirgrabber",
	"webemailextrac",
	"extractorpro",
	"attache",
	"educate search vxb",
	"8484 boston project",
	"franklin locator",
	"nokia-waptoolkit",
	"mailto:craftbot@yahoo.com",
	"full web bot",
	"pcbrowser",
	"psurf",
	"user-Agent",
	"pleasecrawl/1.",
	"kenjin spider",
	"gecko/2525",
	"no browser",
	"webster pro",
	"wep Search 00",
	"grub-client",
	"fastlwspider",
	"this is an exploit",
	"contentsmartz",
	"teleport pro",
	"dts agent",
	"nikto",
	"morzilla",
	"via",
	"atomic_email_hunter",
	"program shareware 1.0.",
	"ecollector",
	"emailcollect",
	"china local browse 2.",
	"backdoor",
	"stress test",
	"foobar/",
	"emailreaper",
	"xmlrpc exploit",
	"compatible ; msie",
	"s.t.a.l.k.e.r.",
	"compatible-",
	"webvulnscan",
	"nameofagent",
	"copyrightcheck",
	"advanced email extractor",
	"surveybot",
	"compatible ;.",
	"searchbot admin@google",
	"wordpress/4.01",
	"webemailextract",
	"larbin@unspecified",
	"turing machine",
	"zeus",
	"windows-update-agent",
	"morfeus fucking scanner",
	"user-agent:",
	"voideye",
	"mosiac 1",
	"chinaclaw",
	"newt activeX; win32",
	"web downloader",
	"safexplorer tl",
	"agdm79@mail.ru",
	"cheesebot",
	"hhjhj@yahoo",
	"fiddler",
	"psycheclone",
	"microsoft internet explorer/5.0",
	"core-project/1",
	"atspider",
	"copyguard",
	"neuralbot/0.2",
	"wordpress hash grabber",
	"amiga-aweb/3.4",
	"packrat",
	"rsync",
	"crescent internet toolpak",
	"security scan",
	"vadixbot",
	"concealed defense",
	"a href=",
	"bwh3_user_agent",
	"internet ninja",
	"microsoft url",
	"emailharvest",
	"shai",
	"wisebot",
	"internet exploiter sux",
	"wells search ii",
	"webroot",
	"digimarc webreader",
	"botversion",
	"black hole",
	"w3mir",
	"pmafind",
	"athens",
	"hl_ftien_spider",
	" injection",
	"takeout",
	"eo browse",
	"cherrypicker",
	"internet-exprorer",
}





func Test_UserAgentIsClean(t *testing.T) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	index := Index( []byte(strings.ToLower(UA)))
	for _,robot := range waf.bad_robots {
		if index.Match(robot) {
			t.Errorf("[%s] found in [%s]", robot, UA )
		}
	}

	if !index.Match([]byte("mozilla/5.0")) {
		t.Errorf("[%s] not found in [%s]", "mozilla/5.0", UA )
	}

	if !waf.BRI_UserAgentIsClean([]byte(strings.ToLower(UA))) {
		t.Errorf("error found in [%s]", UA )
	}


}


func Benchmark_OLD_UserAgentIsClean(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(UA))
	for i := 0; i < b.N; i++ {
		waf.OLD_UserAgentIsClean([]byte(ua))
	}
}

func Benchmark_UserAgentIsClean(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(UA))
	for i := 0; i < b.N; i++ {
		waf.UserAgentIsClean([]byte(ua))
	}
}

func Benchmark_BRI_UserAgentIsClean(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(UA))
	for i := 0; i < b.N; i++ {
		waf.BRI_UserAgentIsClean(ua)
	}
}
