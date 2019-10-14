package transporter

import (
    "bytes"
    "crypto/tls"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    
    goquery "github.com/google/go-querystring/query"
)

const (
    contentType     = "Content-Type"
    jsonContentType = "application/json"
    formContentType = "application/x-www-form-urlencoded"
)

var Tr *http.Transport
var Trs *http.Transport

type Doer interface {
    Do(req *http.Request) (*http.Response, error)
}

type Transporter struct {
    httpClient   Doer
    method       string
    rawURL       string
    header       http.Header
    queryStructs []interface{}
    bodyProvider BodyProvider
}

type SSLConfig struct {
    CAPath  string
    CRTPath string
    KEYPath string
}

func NewTLS(config SSLConfig) (*Transporter, error) {
    
    if config.CRTPath == "" || config.KEYPath == "" {
        return nil, fmt.Errorf("crt and key paths must be provided on TLS connections")
    }
    
    rootCertPool := x509.NewCertPool()
    certs, err := ioutil.ReadDir(config.CAPath)
    if err != nil {
        return nil, err
    }
    
    for _, certF := range certs {
        cert, err := ioutil.ReadFile(config.CAPath + certF.Name())
        if err != nil {
            return nil, err
        }
        rootCertPool.AppendCertsFromPEM(cert)
    }
    
    cert, err := tls.LoadX509KeyPair(config.CRTPath, config.KEYPath)
    if err != nil {
        return nil, fmt.Errorf("client: load keys: %s", err)
    }
    
    var localTrs http.Transport
    
    if Trs == nil {
        Trs = &http.Transport{
            TLSClientConfig: &tls.Config{
                Certificates: []tls.Certificate{cert},
                RootCAs:      rootCertPool,
            },
        }
    }
    
    localTrs = *Trs
    
    client := &http.Client{Transport: &localTrs}
    
    t := &Transporter{
        httpClient:   client,
        method:       "GET",
        header:       make(http.Header),
        queryStructs: make([]interface{}, 0),
    }
    
    return t.Client(client), nil
}

func New() *Transporter {
    
    var localTr http.Transport
    
    if Tr == nil {
        Tr = &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
    }
    
    localTr = *Tr
    
    client := &http.Client{Transport: &localTr}
    
    return &Transporter{
        httpClient:   client,
        method:       "GET",
        header:       make(http.Header),
        queryStructs: make([]interface{}, 0),
    }
    
}

func (t *Transporter) New() *Transporter {
    
    headerCopy := make(http.Header)
    for k, v := range t.header {
        headerCopy[k] = v
    }
    return &Transporter{
        httpClient:   t.httpClient,
        method:       t.method,
        rawURL:       t.rawURL,
        header:       headerCopy,
        queryStructs: append([]interface{}{}, t.queryStructs...),
        bodyProvider: t.bodyProvider,
    }
}

func (t *Transporter) Client(httpClient *http.Client) *Transporter {
    if httpClient == nil {
        return t.Doer(http.DefaultClient)
    }
    return t.Doer(httpClient)
}

func (t *Transporter) Doer(doer Doer) *Transporter {
    if doer == nil {
        t.httpClient = http.DefaultClient
    } else {
        t.httpClient = doer
    }
    return t
}

func (t *Transporter) TLSSet(configssl SSLConfig) (*Transporter, error) {
    
    if configssl.CRTPath != "" && configssl.KEYPath != "" {
        cert, err := tls.LoadX509KeyPair(configssl.CRTPath, configssl.KEYPath)
        if err != nil {
            return nil, fmt.Errorf("client: load keys: %s", err)
        }
        
        config := tls.Config{Certificates: []tls.Certificate{cert}}
        
        tr := &http.Transport{
            TLSClientConfig: &config,
        }
        client := &http.Client{Transport: tr}
        
        return t.Client(client), nil
    } else {
        return t, nil
    }
}

func (t *Transporter) Method(method string) *Transporter {
    t.method = method
    return t
}

func (t *Transporter) Head() *Transporter {
    return t.Method("HEAD")
}

func (t *Transporter) Get() *Transporter {
    return t.Method("GET")
}

func (t *Transporter) Post() *Transporter {
    return t.Method("POST")
}

func (t *Transporter) Put() *Transporter {
    return t.Method("PUT")
}

func (t *Transporter) Patch() *Transporter {
    return t.Method("PATCH")
}

func (t *Transporter) Delete() *Transporter {
    return t.Method("DELETE")
}

func (t *Transporter) Add(key, value string) *Transporter {
    t.header.Add(key, value)
    return t
}

func (t *Transporter) Set(key, value string) *Transporter {
    t.header.Set(key, value)
    return t
}

func (t *Transporter) SetHeader(h *http.Header) *Transporter {
    t.header = *h
    return t
}

func (t *Transporter) SetBasicAuth(username, password string) *Transporter {
    return t.Set("Authorization", "Basic "+basicAuth(username, password))
}

func (t *Transporter) SetJwtAuth(jwt string) *Transporter {
    return t.Set("Authorization", "jwt "+jwt)
}

func (t *Transporter) SetOAuth(oauth string) *Transporter {
    return t.Set("Authorization", "Bearer "+oauth)
}

func basicAuth(username, password string) string {
    auth := username + ":" + password
    return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (t *Transporter) Base(rawURL string) *Transporter {
    t.rawURL = rawURL
    return t
}

func (t *Transporter) Path(path string) *Transporter {
    baseURL, baseErr := url.Parse(t.rawURL)
    pathURL, pathErr := url.Parse(path)
    if baseErr == nil && pathErr == nil {
        t.rawURL = baseURL.ResolveReference(pathURL).String()
        return t
    }
    return t
}

func (t *Transporter) QueryStruct(queryStruct interface{}) *Transporter {
    if queryStruct != nil {
        t.queryStructs = append(t.queryStructs, queryStruct)
    }
    return t
}

func (t *Transporter) Body(body io.Reader) *Transporter {
    if body == nil {
        return t
    }
    return t.BodyProvider(bodyProvider{body: body})
}

func (t *Transporter) BodyProvider(body BodyProvider) *Transporter {
    if body == nil {
        return t
    }
    t.bodyProvider = body
    
    ct := body.ContentType()
    if ct != "" {
        t.Set(contentType, ct)
    }
    
    return t
}

func (t *Transporter) BodyJSON(bodyJSON interface{}) *Transporter {
    if bodyJSON == nil {
        return t
    }
    return t.BodyProvider(jsonBodyProvider{payload: bodyJSON})
}

func (t *Transporter) BodyForm(bodyForm interface{}) *Transporter {
    if bodyForm == nil {
        return t
    }
    return t.BodyProvider(formBodyProvider{payload: bodyForm})
}

func (t *Transporter) ReceiveSuccess(successV interface{}) (*http.Response, error) {
    return t.Receive(successV, nil)
}

func (t *Transporter) Receive(successV, failureV interface{}) (*http.Response, error) {
    req, err := t.Request()
    if err != nil {
        return nil, err
    }
    return t.Do(req, successV, failureV)
}

func (t *Transporter) Request() (*http.Request, error) {
    reqURL, err := url.Parse(t.rawURL)
    if err != nil {
        return nil, err
    }
    
    err = addQueryStructs(reqURL, t.queryStructs)
    if err != nil {
        return nil, err
    }
    
    var body io.Reader
    if t.bodyProvider != nil {
        body, err = t.bodyProvider.Body()
        if err != nil {
            return nil, err
        }
    }
    req, err := http.NewRequest(t.method, reqURL.String(), body)
    if err != nil {
        return nil, err
    }
    addHeaders(req, t.header)
    return req, err
}

func addQueryStructs(reqURL *url.URL, queryStructs []interface{}) error {
    urlValues, err := url.ParseQuery(reqURL.RawQuery)
    if err != nil {
        return err
    }
    
    for _, queryStruct := range queryStructs {
        queryValues, err := goquery.Values(queryStruct)
        if err != nil {
            return err
        }
        for key, values := range queryValues {
            for _, value := range values {
                urlValues.Add(key, value)
            }
        }
    }
    
    reqURL.RawQuery = urlValues.Encode()
    return nil
}

func addHeaders(req *http.Request, header http.Header) {
    for key, values := range header {
        for _, value := range values {
            req.Header.Add(key, value)
        }
    }
}

func (t *Transporter) Do(req *http.Request, successV, failureV interface{}) (*http.Response, error) {
    resp, err := t.httpClient.Do(req)
    if err != nil {
        return resp, err
    }
    
    defer resp.Body.Close()
    
    if resp.StatusCode == 204 {
        return resp, nil
    }
    
    if successV != nil || failureV != nil {
        err = decodeResponse(resp, successV, failureV)
    }
    return resp, err
}

func decodeText(content []byte, v interface{}) error {
    sv := v.(*string)
    *sv = string(content)
    return nil
}

func decodeJSON(content []byte, v interface{}) error {
    return json.Unmarshal(content, v)
}

func decodeResponse(resp *http.Response, successV, failureV interface{}) error {
    content, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    if code := resp.StatusCode; 200 <= code && code <= 299 {
        switch successV.(type) {
        case *string:
            err = decodeText(content, successV)
        case nil:
            err = nil
        default:
            err = decodeJSON(content, successV)
        }
    } else {
        switch failureV.(type) {
        case *string:
            err = decodeText(content, failureV)
        case nil:
            err = nil
        default:
            err = decodeJSON(content, failureV)
        }
    }
    if err != nil {
        err = fmt.Errorf("%s\nhappend while parsing: %s", err.Error(), content)
    }
    return err
}

type BodyProvider interface {
    ContentType() string
    Body() (io.Reader, error)
}

type bodyProvider struct {
    body io.Reader
}

type jsonBodyProvider struct {
    payload interface{}
}

type formBodyProvider struct {
    payload interface{}
}

func (p bodyProvider) ContentType() string {
    return ""
}

func (p bodyProvider) Body() (io.Reader, error) {
    return p.body, nil
}

func (p jsonBodyProvider) ContentType() string {
    return jsonContentType
}

func (p jsonBodyProvider) Body() (io.Reader, error) {
    buf := &bytes.Buffer{}
    err := json.NewEncoder(buf).Encode(p.payload)
    if err != nil {
        return nil, err
    }
    return buf, nil
}

func (p formBodyProvider) ContentType() string {
    return formContentType
}

func (p formBodyProvider) Body() (io.Reader, error) {
    values, err := goquery.Values(p.payload)
    if err != nil {
        return nil, err
    }
    return strings.NewReader(values.Encode()), nil
}
