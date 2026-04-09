package model

// URL is a string representing a Uniform Resource Locator (URL).
// example: https://api.ByteLyon.com:8080/bots?type=news#latest
// scheme: https://
// host: api.ByteLyon.com
// subdomain: api
// domain: ByteLyon.com
// port: 8080
// path: /bots?type=news
// fragment: latest
type URL string
