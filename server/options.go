package server

type Options struct {
	EnableRandomUrl bool `hcl:"enable_random_url" flagName:"random-url" flagSName:"r" flagDescribe:"Add a random string to the URL" default:"false"`
}
