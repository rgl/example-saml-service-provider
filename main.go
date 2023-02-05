package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/crewjam/saml/samlsp"
)

var indexTextTemplate = template.Must(template.New("Index").Parse(`# SAML Attributes
{{- range $kv := .SAMLAttributes}}
{{- range .Values }}
{{$kv.Name}}: {{.}}
{{- end}}
{{- end}}
`))

var indexTemplate = template.Must(template.New("Index").Parse(`<!DOCTYPE html>
<html>
<head>
<title>example-saml-service-provider</title>
<style>
body {
	font-family: monospace;
	color: #555;
	background: #e6edf4;
	padding: 1.25rem;
	margin: 0;
}
table {
	background: #fff;
	border: .0625rem solid #c4cdda;
	border-radius: 0 0 .25rem .25rem;
	border-spacing: 0;
	margin-bottom: 1.25rem;
	padding: .75rem 1.25rem;
	text-align: left;
	white-space: pre;
}
table > caption {
	background: #f1f6fb;
	text-align: left;
	font-weight: bold;
	padding: .75rem 1.25rem;
	border: .0625rem solid #c4cdda;
	border-radius: .25rem .25rem 0 0;
	border-bottom: 0;
}
table td, table th {
	padding: .25rem;
}
table > tbody > tr:hover {
	background: #f1f6fb;
}
</style>
</head>
<body>
	<table>
		<caption>Actions</caption>
		<tbody>
			<tr><td><a href="/login">login</a></td></tr>
		</tbody>
	</table>
	{{- if .SAMLAttributes }}
	<table>
		<caption>SAML Attributes</caption>
		<tbody>
			{{- range $kv := .SAMLAttributes }}
			{{- range .Values }}
			<tr>
				<th>{{$kv.Name}}</th>
				<td>{{.}}</td>
			</tr>
			{{- end}}
			{{- end}}
		</tbody>
	</table>
	{{- end}}
</body>
</html>
`))

type keyValue struct {
	Name   string
	Values []string
}

type keyValues []keyValue

func (a keyValues) Len() int      { return len(a) }
func (a keyValues) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a keyValues) Less(i, j int) bool {
	return strings.ToLower(a[i].Name) < strings.ToLower(a[j].Name)
}

type indexData struct {
	SAMLAttributes keyValues
}

func getSAMLAttributes(s samlsp.Session) keyValues {
	if s == nil {
		return keyValues{}
	}
	sa, ok := s.(samlsp.SessionWithAttributes)
	if !ok {
		return keyValues{}
	}
	samlAttributes := sa.GetAttributes()
	result := make(keyValues, 0, len(samlAttributes))
	for k := range samlAttributes {
		result = append(result, keyValue{
			Name:   k,
			Values: samlAttributes[k],
		})
	}
	sort.Sort(result)
	return result
}

func index(w http.ResponseWriter, r *http.Request) {
	s := samlsp.SessionFromContext(r.Context())
	samlAttributes := getSAMLAttributes(s)

	var t *template.Template
	var contentType string

	switch r.URL.Query().Get("format") {
	case "text":
		t = indexTextTemplate
		contentType = "text/plain"
	default:
		t = indexTemplate
		contentType = "text/html"
	}

	w.Header().Set("Content-Type", contentType)

	err := t.ExecuteTemplate(w, "Index", indexData{
		SAMLAttributes: samlAttributes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	log.SetFlags(0)

	var listenFlag = flag.String("listen", "http://localhost:8000", "Listen URL")
	var entityIDFlag = flag.String("entity-id", "urn:example:example-saml-service-provider", "Service provider Entity ID")
	var idpMetadataFlag = flag.String("idp-metadata", "https://samltest.id/saml/idp", "IDP Metadata URL")

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("\nERROR You MUST NOT pass any positional arguments")
	}

	keyPair, err := tls.LoadX509KeyPair(
		"example-saml-service-provider-crt.pem",
		"example-saml-service-provider-key.pem")
	if err != nil {
		log.Panicf("Failed to load the service provider key pair: %v", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		log.Panicf("Failed to parse the service provider certificate: %v", err)
	}

	idpMetadataURL, err := url.Parse(*idpMetadataFlag)
	if err != nil {
		log.Panicf("Failed to parse the IDP metadata url: %v", err)
	}
	idpMetadata, err := samlsp.FetchMetadata(
		context.Background(),
		http.DefaultClient,
		*idpMetadataURL)
	if err != nil {
		log.Panicf("Failed to fetch the IDP metadata: %v", err)
	}

	rootURL, err := url.Parse(*listenFlag)
	if err != nil {
		log.Panicf("Failed to parse the service provider url: %v", err)
	}

	samlMiddleware, _ := samlsp.New(samlsp.Options{
		EntityID:          *entityIDFlag,
		URL:               *rootURL,
		Key:               keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:       keyPair.Leaf,
		IDPMetadata:       idpMetadata,
		AllowIDPInitiated: true,
	})

	buf, err := xml.MarshalIndent(samlMiddleware.ServiceProvider.Metadata(), "", "  ")
	if err != nil {
		log.Panicf("Failed to marshal the service provider metadata: %v", err)
	}
	err = os.WriteFile("example-saml-service-provider-metadata.xml", buf, 0664)
	if err != nil {
		log.Printf("Warning: failed to save the service provider metadata to local file: %v", err)
	}

	http.Handle("/", http.HandlerFunc(index))
	http.Handle("/login", samlMiddleware.RequireAccount(http.HandlerFunc(index)))
	http.Handle("/saml/", samlMiddleware)
	log.Printf("Service provider listening at %s", rootURL)
	http.ListenAndServe(":"+rootURL.Port(), nil)
}
