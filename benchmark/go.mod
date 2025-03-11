module github.com/semihalev/twig/benchmark

go 1.24.1

require (
	github.com/flosch/pongo2/v6 v6.0.0
	github.com/semihalev/twig v0.0.0
	github.com/tyler-sommer/stick v1.0.6
	github.com/valyala/quicktemplate v1.8.0
)

replace github.com/semihalev/twig => ../

require (
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
)
