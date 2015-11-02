go-saml
======

[![Build Status](https://travis-ci.org/RobotsAndPencils/go-saml.svg?branch=master)](https://travis-ci.org/RobotsAndPencils/go-saml)

A just good enough SAML client library written in Go. This library is by no means complete and has been developed
to solve several specific integration efforts. However, it's a start, and it would be great to see
it evolve into a more fleshed out implemention.

Inspired by the early work of [Matt Baird](https://github.com/mattbaird/gosaml).

The library supports:

* generating signed/unsigned AuthnRequests
* validating signed AuthnRequests
* generating service provider metadata
* generating signed Responses
* validating signed Responses


Installation
------------

    $ go get github.com/RobotsAndPencils/go-saml

Here's a convenient way to generate a certificate:

    curl -sSL https://raw.githubusercontent.com/frntn/x509-san/master/gencert.sh | CRT_CN="mycert"  bash


Usage
-----

Below are samples to show how you might use the library.

### Generating Signed AuthnRequests

```go
sp := saml.ServiceProviderSettings{
  PublicCertPath:              "../default.crt",
  PrivateKeyPath:              "../default.key",
  IDPSSOURL:                   "http://idp/saml2",
  IDPSSODescriptorURL:         "http://idp/issuer",
  IDPPublicCertPath:           "idpcert.crt",
  SPSignRequest:               "true",
  AssertionConsumerServiceURL: "http://localhost:8000/saml_consume",
}
sp.Init()

// generate the AuthnRequest and then get a base64 encoded string of the XML
authnRequest := sp.GetAuthnRequest()
b64XML, err := authnRequest.EncodedSignedString(sp.PrivateKeyPath)
if err != nil {
  panic(err)
}

// for convenience, get a URL formed with the SAMLRequest parameter
url, err := saml.GetAuthnRequestURL(sp.IDPSSOURL, b64XML)
if err != nil {
  panic(err)
}

// below is bonus for how you might respond to a request with a form that POSTs to the IdP
data := struct {
  Base64AuthRequest string
  URL               string
}{
  Base64AuthRequest: b64XML,
  URL:               url,
}

t := template.New("saml")
t, err = t.Parse("<html><body style=\"display: none\" onload=\"document.frm.submit()\"><form method=\"post\" name=\"frm\" action=\"{{.URL}}\"><input type=\"hidden\" name=\"SAMLRequest\" value=\"{{.Base64AuthRequest}}\" /><input type=\"submit\" value=\"Submit\" /></form></body></html>")

// how you might respond to a request with the templated form that will auto post
t.Execute(w, data)
```

### Validating a received SAML Response


```go
response = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  encodedXML := r.FormValue("SAMLResponse")

  if encodedXML == "" {
    httpcommon.SendBadRequest(w, "SAMLResponse form value missing")
    return
  }

  response, err := saml.ParseEncodedResponse(encodedXML)
  if err != nil {
    httpcommon.SendBadRequest(w, "SAMLResponse parse: "+err.Error())
    return
  }

  err = response.Validate(&sp)
  if err != nil {
    httpcommon.SendBadRequest(w, "SAMLResponse validation: "+err.Error())
    return
  }

  samlID := response.GetAttribute("uid")
  if samlID == "" {
    httpcommon.SendBadRequest(w, "SAML attribute identifier uid missing")
    return
  }

  //...
}
```

### Service provider metadata

```go
func samlMetadataHandler(sp *saml.ServiceProviderSettings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md, err := sp.GetEntityDescriptor()
		if err != nil {
      w.WriteHeader(500)
      w.Write([]byte("Error: " + err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(md))
	})
}
```

### Receiving a authnRequest

```go
b64Request := r.URL.Query().Get("SAMLRequest")
if b64Request == "" {
  w.WriteHeader(400)
  w.Write([]byte("SAMLRequest parameter missing"))
  return
}

defated, err := base64.StdEncoding.DecodeString(b64Request)
if err != nil {
  w.WriteHeader(500)
  w.Write([]byte("Error: " + err.Error()))
  return
}

// enflate and unmarshal
var buffer bytes.Buffer
rdr := flate.NewReader(bytes.NewReader(defated))
io.Copy(&buffer, rdr)
var authnRequest saml.AuthnRequest

err = xml.Unmarshal(buffer.Bytes(), &authnRequest)
if err != nil {
  w.WriteHeader(500)
  w.Write([]byte("Error: " + err.Error()))
  return
}

if authnRequest.Issuer.Url != issuerURL {
  w.WriteHeader(500)
  w.Write([]byte("unauthorized issuer "+authnRequest.Issuer.Url))
  return
}

```

### Creating a SAML Response (if acting as an IdP)

```go
issuer := "http://localhost:8000/saml"
authnResponse := saml.NewSignedResponse()
authnResponse.Issuer.Url = issuer
authnResponse.Assertion.Issuer.Url = issuer
authnResponse.Signature.KeyInfo.X509Data.X509Certificate.Cert = stringValueOfCert
authnResponse.Assertion.Subject.NameID.Value = userIdThatYouAuthenticated
authnResponse.AddAttribute("uid", userIdThatYouAuthenticated)
authnResponse.AddAttribute("email", "someone@domain")
authnResponse.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo = authnRequestIdRespondingTo
authnResponse.InResponseTo = authnRequestIdRespondingTo
authnResponse.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.Recipient = issuer

// signed XML string
signed, err := authnResponse.SignedString("/path/to/private.key")

// or signed base64 encoded XML string
b64XML, err := authnResponse.EncodedSignedString("/path/to/private.key")

```


### Contributing

Would love any contributions you having including better documentation, tests, or more robust functionality.

    git clone git@github.com:RobotsAndPencils/go-saml.git
    make init
    make test

### Contact

[![Robots & Pencils Logo](http://f.cl.ly/items/2W3n1r2R0j2p2b3n3j3c/rnplogo.png)](http://www.robotsandpencils.com)

Made with :heart: by Robots & Pencils ([@robotsNpencils](https://twitter.com/robotsNpencils))

#### Maintainers

- [Mike Brevoort](http://github.com/mbrevoort) ([@mbrevoort](https://twitter.com/mbrevoort))
