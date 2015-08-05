package waf


import	(
	"testing"
	"strings"
)




var UA		= "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:40.0) Gecko/20100101 Firefox/40.0"
var BadUAbeg	= "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:40.0) email Gecko/20100101 Firefox/40.0"
var BadUAend	= "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:40.0) wordpress Gecko/20100101 Firefox/40.0"


var loremipsum = `
	Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam auctor, nisl cursus vestibulum dapibus, mi massa cursus massa, quis mattis eros est et purus. Proin congue consequat tellus a bibendum. Quisque varius nunc vel massa aliquet, vitae tincidunt mauris lacinia. Fusce metus quam, consequat sit amet blandit sed, finibus eu metus. Donec mattis leo sed sapien dictum commodo. Maecenas lorem nisl, feugiat at vehicula eget, porta vitae nisi. Vestibulum dapibus ornare tellus et finibus. Etiam volutpat orci quis erat condimentum, sed condimentum purus tempus. Nulla a ante placerat, scelerisque massa ac, laoreet dui. Suspendisse potenti. Sed semper leo orci, sit amet convallis lectus luctus id. Ut ac pellentesque nibh. Suspendisse odio magna, bibendum ut laoreet sed, auctor nec libero. Pellentesque vel felis arcu. Nunc sed sapien in odio ultrices imperdiet. Donec ac nunc id felis posuere scelerisque sit amet vel magna.
	Vestibulum ornare dolor ut risus scelerisque imperdiet. Maecenas sed turpis tortor. Integer id aliquam metus, vel varius leo. Nulla sit amet nunc ut arcu tempus molestie. Quisque dapibus, elit quis vestibulum vehicula, lectus justo suscipit quam, sed accumsan sem arcu eget velit. Donec tincidunt, orci a eleifend finibus, lectus turpis pulvinar sem, sit amet fringilla est odio et nisl. Phasellus in lacinia metus, vel accumsan diam. Cras luctus id augue ac consectetur. Mauris molestie nec nunc ac sodales.
	Suspendisse tempus, nisl a fringilla dapibus, ex nisi malesuada quam, posuere suscipit purus nisi et sem. Fusce elementum ante sed ante auctor, quis rutrum est viverra. Nam tincidunt ultricies nulla nec condimentum. Aenean eu maximus arcu, vitae elementum orci. Donec at sem sagittis, congue sem eu, gravida urna. Nulla ac lacus suscipit, convallis risus a, tincidunt mi. Curabitur ultricies rutrum lectus, vel viverra massa. Mauris ullamcorper, urna a tempor sagittis, leo lectus fringilla magna, et laoreet nisl ligula a sapien. Ut eu gravida est, at lobortis lorem. Duis iaculis lorem nec velit eleifend, eget imperdiet magna gravida. Nunc porta mauris lectus, eget consectetur nulla mollis sed. Proin volutpat velit a metus imperdiet, a varius risus cursus. Integer auctor tincidunt semper. Maecenas rutrum tellus non fermentum viverra. Phasellus venenatis est ut mollis tincidunt.
	Sed vestibulum neque vitae scelerisque sagittis. Morbi elementum maximus pharetra. Sed eu consequat nisi. Morbi imperdiet nibh nec est tincidunt bibendum. Sed cursus sit amet justo eget congue. Etiam ullamcorper tellus metus, quis elementum risus mollis quis. Pellentesque sit amet risus vitae mi pellentesque posuere. Praesent lacinia nisi erat, vel tempor nibh lacinia finibus. Maecenas sed blandit sem, et rhoncus urna. Praesent laoreet ornare purus, in cursus sem facilisis vitae. Curabitur mollis sagittis convallis. Quisque sollicitudin porta dictum. Aliquam erat volutpat. Aenean ornare tincidunt ante, a porttitor lectus varius quis. Aliquam erat volutpat. Donec tempor purus a felis aliquam dapibus.
	Phasellus porttitor venenatis neque, quis tempor lorem lobortis at. Fusce quis dolor ut leo hendrerit molestie a sit amet mauris. Sed vitae convallis arcu. Quisque urna lectus, placerat non arcu maximus, maximus fringilla neque. Nullam facilisis ex non lectus lacinia gravida. Ut eu libero dignissim, vulputate massa eget, aliquam est. Vivamus malesuada orci ac elit consectetur tempus. Aenean varius euismod urna, eu faucibus sem sodales vel.
	Praesent in nisl tempor, elementum eros ut, volutpat erat. Aliquam molestie, tellus sed tempor rutrum, turpis risus efficitur turpis, sed hendrerit purus sapien in dolor. Quisque malesuada tempus lacus, et lacinia metus feugiat eu. Vivamus sollicitudin ipsum sed mi posuere suscipit. Mauris ut lobortis sem. In at eleifend urna, at accumsan sem. In ac lacus ac augue feugiat mollis. Phasellus viverra rhoncus ornare. Suspendisse potenti. Morbi nec egestas metus. Nam elementum posuere diam, ac rutrum elit pellentesque ac. Nullam lacinia quam ac lectus porta dapibus. Ut aliquam massa non tellus lobortis, sed rutrum est ultrices.
	Curabitur auctor et ipsum in ultricies. Aliquam erat volutpat. Aliquam erat volutpat. Cras ut quam maximus urna molestie fermentum eget vel turpis. Donec ac auctor elit, non tincidunt ipsum. Interdum et malesuada fames ac ante ipsum primis in faucibus. Quisque vitae metus eu turpis tristique tristique. Fusce non tellus mi. Ut arcu ipsum, pulvinar eu sem eget, fermentum lacinia lectus. Duis ut pretium nisi. Vivamus scelerisque hendrerit metus, sed gravida nisl commodo nec. Proin laoreet nulla dui, efficitur feugiat diam rhoncus sodales. Nulla nunc felis, consequat in mattis sed, bibendum et sapien. Vivamus eu luctus ante, ut mattis risus. Maecenas aliquam feugiat nulla, at porttitor ante.
	Maecenas eget mauris blandit, vulputate justo sit amet, gravida magna. Aliquam molestie lectus a libero condimentum, sollicitudin feugiat sapien pharetra. Ut fermentum pulvinar dolor id feugiat. Pellentesque eget risus eget ipsum finibus feugiat in a massa. Aenean non sapien at dui sagittis vestibulum. Quisque pretium varius pellentesque. Nullam eu sapien euismod, mollis ipsum vel, placerat lacus. Vestibulum dictum laoreet lorem, nec bibendum elit placerat vitae. Aenean tellus arcu, eleifend sed ipsum ac, aliquam rutrum ipsum. Morbi ligula velit, maximus sit amet condimentum eu, dapibus ut sapien. Fusce et venenatis orci, sed vehicula orci. `



var loremipsumBAD = `
	Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam auctor, nisl cursus vestibulum dapibus, mi massa cursus massa, quis mattis eros est et purus. Proin congue consequat tellus a bibendum. Quisque varius nunc vel massa aliquet, vitae tincidunt mauris lacinia. Fusce metus quam, consequat sit amet blandit sed, finibus eu metus. Donec mattis leo sed sapien dictum commodo. Maecenas lorem nisl, feugiat at vehicula eget, porta vitae nisi. Vestibulum dapibus ornare tellus et finibus. Etiam volutpat orci quis erat condimentum, sed condimentum purus tempus. Nulla a ante placerat, scelerisque massa ac, laoreet dui. Suspendisse potenti. Sed semper leo orci, sit amet convallis lectus luctus id. Ut ac pellentesque nibh. Suspendisse odio magna, bibendum ut laoreet sed, auctor nec libero. Pellentesque vel felis arcu. Nunc sed sapien in odio ultrices imperdiet. Donec ac nunc id felis posuere scelerisque sit amet vel magna.
	Vestibulum ornare dolor ut risus scelerisque imperdiet. Maecenas sed turpis tortor. Integer id aliquam metus, vel varius leo. Nulla sit amet nunc ut arcu tempus molestie. Quisque dapibus, elit quis vestibulum vehicula, lectus justo suscipit quam, sed accumsan sem arcu eget velit. Donec tincidunt, orci a eleifend finibus, lectus turpis pulvinar sem, sit amet fringilla est odio et nisl. Phasellus in lacinia metus, vel accumsan diam. Cras luctus id augue ac consectetur. Mauris molestie nec nunc ac sodales.
	Suspendisse tempus, nisl a fringilla dapibus, ex nisi malesuada quam, posuere suscipit purus nisi et sem. Fusce elementum ante sed ante auctor, quis rutrum est viverra. Nam tincidunt ultricies nulla nec condimentum. Aenean eu maximus arcu, vitae elementum orci. Donec at sem sagittis, congue sem eu, gravida urna. Nulla ac lacus suscipit, convallis risus a, tincidunt mi. Curabitur ultricies rutrum lectus, vel viverra massa. Mauris ullamcorper, urna a tempor sagittis, leo lectus fringilla magna, et laoreet nisl ligula a sapien. Ut eu gravida est, at lobortis lorem. Duis iaculis lorem nec velit eleifend, eget imperdiet magna gravida. Nunc porta mauris lectus, eget consectetur nulla mollis sed. Proin volutpat velit a metus imperdiet, a varius risus cursus. Integer auctor tincidunt semper. Maecenas rutrum tellus non fermentum viverra. Phasellus venenatis est ut mollis tincidunt.
	Sed vestibulum neque vitae scelerisque sagittis. Morbi elementum maximus pharetra. Sed eu consequat nisi. Morbi imperdiet nibh nec est tincidunt bibendum. Sed cursus sit amet justo eget congue. Etiam ullamcorper tellus metus, quis elementum risus mollis quis. Pellentesque sit amet risus vitae mi pellentesque posuere. Praesent lacinia nisi erat, vel tempor nibh lacinia finibus. Maecenas sed blandit sem, et rhoncus urna. Praesent laoreet ornare purus, in cursus sem facilisis vitae. Curabitur mollis sagittis convallis. Quisque sollicitudin porta dictum. Aliquam erat volutpat. Aenean ornare tincidunt ante, a porttitor lectus varius quis. Aliquam erat volutpat. Donec tempor purus a felis aliquam dapibus.
	Phasellus porttitor venenatis neque, quis tempor lorem lobortis at. Fusce quis dolor ut leo hendrerit molestie a sit amet mauris. Sed vitae convallis arcu. Quisque urna lectus, placerat non arcu maximus, maximus fringilla neque. Nullam facilisis ex non lectus lacinia gravida. Ut eu libero dignissim, vulputate massa eget, aliquam est. Vivamus malesuada orci ac elit consectetur tempus. Aenean varius euismod urna, eu faucibus sem sodales vel.
	Praesent in nisl tempor, elementum eros ut, volutpat erat. Aliquam molestie, tellus sed tempor rutrum, turpis risus efficitur turpis, sed hendrerit purus sapien in dolor. Quisque malesuada tempus lacus, et lacinia metus feugiat eu. Vivamus sollicitudin ipsum sed mi posuere suscipit. Mauris ut lobortis sem. In at eleifend urna, at accumsan sem. In ac lacus ac augue feugiat mollis. Phasellus viverra rhoncus ornare. Suspendisse potenti. Morbi nec egestas metus. Nam elementum posuere diam, ac rutrum elit pellentesque ac. Nullam lacinia quam ac lectus porta dapibus. Ut aliquam massa non tellus lobortis, sed rutrum est ultrices. autoemailspider
	Curabitur auctor et ipsum in ultricies. Aliquam erat volutpat. Aliquam erat volutpat. Cras ut quam maximus urna molestie fermentum eget vel turpis. Donec ac auctor elit, non tincidunt ipsum. Interdum et malesuada fames ac ante ipsum primis in faucibus. Quisque vitae metus eu turpis tristique tristique. Fusce non tellus mi. Ut arcu ipsum, pulvinar eu sem eget, fermentum lacinia lectus. Duis ut pretium nisi. Vivamus scelerisque hendrerit metus, sed gravida nisl commodo nec. Proin laoreet nulla dui, efficitur feugiat diam rhoncus sodales. Nulla nunc felis, consequat in mattis sed, bibendum et sapien. Vivamus eu luctus ante, ut mattis risus. Maecenas aliquam feugiat nulla, at porttitor ante.
	Maecenas eget mauris blandit, vulputate justo sit amet, gravida magna. Aliquam molestie lectus a libero condimentum, sollicitudin feugiat sapien pharetra. Ut fermentum pulvinar dolor id feugiat. Pellentesque eget risus eget ipsum finibus feugiat in a massa. Aenean non sapien at dui sagittis vestibulum. Quisque pretium varius pellentesque. Nullam eu sapien euismod, mollis ipsum vel, placerat lacus. Vestibulum dictum laoreet lorem, nec bibendum elit placerat vitae. Aenean tellus arcu, eleifend sed ipsum ac, aliquam rutrum ipsum. Morbi ligula velit, maximus sit amet condimentum eu, dapibus ut sapien. Fusce et venenatis orci, sed vehicula orci. `



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
	"grabber",
	"cgichk",
	"bsqlbf",
	"mozilla/4.0 (compatible)",
	"sqlmap",
	"mozilla/4.0 (compatible; msie 6.0; win32)",
	"mozilla/5.0 sf//",
	"nessus",
	"arachni",
	"metis",
	"sql power injector",
	"bilbo",
	"absinthe",
	"black widow",
	"n-stealth",
	"brutus",
	"webtrends security analyzer",
	"netsparker",
	"python-httplib2",
	"jaascois",
	"pmafind",
	".nasl",
	"nsauditor",
	"paros",
	"dirbuster",
	"pangolin",
	"nmap nse",
	"sqlninja",
	"nikto",
	"webinspect",
	"blackwidow",
	"grendel-scan",
	"havij",
	"w3af",
	"hydra",
	"super happy fun",
	"psycheclone",
	"grub crawler",
	"core-project/",
	"winnie poh",
	"mozilla/4.0+(",
	"email siphon",
	"internet explorer",
	"nutscrape/",
	"mozilla/4.0(",
	"missigua",
	"libwww-perl",
	"movable type",
	"user",
	"blogsearchbot-martin",
	"emailsiphon",
	"digger",
	"8484 boston project",
	"nutchcvs",
	"pycurl",
	"java 1.",
	"isc systems irc",
	"emailcollector",
	"mj12bot/v1.0.8",
	"trackback/",
	"microsoft url",
	"diamond",
	"autoemailspider",
	"lwp",
	"pussycat",
	"jakarta commons",
	"java/1.",
	"user-agent:",
	"<sc",
	"adwords",
	"omniexplorer",
	"wordpress",
	"httpproxy",
	"user agent:",
	"ecollector",
	"msie",
	"cherrypicker",







	"zoubidatralala",



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



	if !waf.GoSufArray_UserAgentIsClean([]byte(strings.ToLower(UA))) {
		t.Errorf("error found in GSA [%s]", UA )
	}

	if waf.GoSufArray_UserAgentIsClean([]byte(strings.ToLower(BadUAbeg))) {
		t.Errorf("error found in GSA [%s]", BadUAbeg )
	}

	if waf.GoSufArray_UserAgentIsClean([]byte(strings.ToLower(BadUAend))) {
		t.Errorf("error found in GSA [%s]", BadUAend )
	}


	if !waf.BRS_UserAgentIsClean([]byte(strings.ToLower(UA))) {
		t.Errorf("error found in BRS [%s]", UA )
	}

	if waf.BRS_UserAgentIsClean([]byte(strings.ToLower(BadUAbeg))) {
		t.Errorf("error found in BRS [%s]", BadUAbeg )
	}

	if waf.BRS_UserAgentIsClean([]byte(strings.ToLower(BadUAend))) {
		t.Errorf("error found in BRS [%s]", BadUAend )
	}


	if !waf.BRS_UserAgentIsClean([]byte(strings.ToLower(loremipsum))) {
		t.Errorf("error found in BRS [LoremIpsum]" )
	}

	if waf.BRS_UserAgentIsClean([]byte(strings.ToLower(loremipsumBAD))) {
		t.Errorf("error found in BRS [BADLoremIpsum]" )
	}



}




func Benchmark_BRS_UserAgentIsClean_begInvUA(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(BadUAbeg))
	for i := 0; i < b.N; i++ {
		waf.BRS_UserAgentIsClean(ua)
	}
}

func Benchmark_BRS_UserAgentIsClean_endInvUA(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(BadUAend))
	for i := 0; i < b.N; i++ {
		waf.BRS_UserAgentIsClean(ua)
	}
}

func Benchmark_BRS_UserAgentIsClean_validUA(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(UA))
	for i := 0; i < b.N; i++ {
		waf.BRS_UserAgentIsClean(ua)
	}
}



func Benchmark_BRS_UserAgentIsClean_OKLorem(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(loremipsum))
	for i := 0; i < b.N; i++ {
		waf.BRS_UserAgentIsClean(ua)
	}
}

func Benchmark_BRS_UserAgentIsClean_BadLorem(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(loremipsumBAD))
	for i := 0; i < b.N; i++ {
		waf.BRS_UserAgentIsClean(ua)
	}
}






func Benchmark_GSA_UserAgentIsClean_validUA(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(UA))
	for i := 0; i < b.N; i++ {
		waf.GoSufArray_UserAgentIsClean([]byte(ua))
	}
}

func Benchmark_GSA_UserAgentIsClean_begInvUA(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(BadUAbeg))
	for i := 0; i < b.N; i++ {
		waf.GoSufArray_UserAgentIsClean([]byte(ua))
	}
}

func Benchmark_GSA_UserAgentIsClean_endInvUA(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(BadUAend))
	for i := 0; i < b.N; i++ {
		waf.GoSufArray_UserAgentIsClean([]byte(ua))
	}
}



func Benchmark_GSA_UserAgentIsClean_OKlorem(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(loremipsum))
	for i := 0; i < b.N; i++ {
		waf.GoSufArray_UserAgentIsClean([]byte(ua))
	}
}

func Benchmark_GSA_UserAgentIsClean_badLorem(b *testing.B) {
	waf := new(WAF)
	waf.load_bad_robots(bad_robots)

	ua := []byte(strings.ToLower(loremipsumBAD))
	for i := 0; i < b.N; i++ {
		waf.GoSufArray_UserAgentIsClean([]byte(ua))
	}
}
