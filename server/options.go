package server

type Options struct {
	EnableRandomUrl bool   `hcl:"enable_random_url" flagName:"random-url" flagSName:"r" flagDescribe:"Add a random string to the URL" default:"false"`
	RandomUrlLength int    `hcl:"random_url_length" flagName:"random-url-length" flagDescribe:"Random URL length" default:"8"`
	TitleFormat     string `hcl:"title_format" flagName:"title-format" flagSName:"" flagDescribe:"Title format of browser window" default:"{{ .command }}@{{ .hostname }}"`
	PermitArguments bool   `hcl:"permit_arguments" flagName:"permit-arguments" flagDescribe:"Permit clients to send command line arguments in URL (e.g. http://example.com:8080/?arg=AAA&arg=BBB)" default:"true"`

	TitleVariables map[string]interface{}
}
