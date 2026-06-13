module github.com/SirNiklas9/projx-workflow

go 1.25.0

require (
	github.com/SirNiklas9/projx-context v0.0.0
	github.com/SirNiklas9/projx-core v0.0.0
	github.com/SirNiklas9/projx-generation v0.0.0
	github.com/SirNiklas9/projx-store v0.0.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/odvcencio/gotreesitter v0.20.2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/sys v0.42.0 // indirect
	modernc.org/libc v1.72.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.52.0 // indirect
)

// Local development: separate repos, co-developed.
replace github.com/SirNiklas9/projx-core => ../projx-core

replace github.com/SirNiklas9/projx-context => ../projx-context

replace github.com/SirNiklas9/projx-generation => ../projx-generation

replace github.com/SirNiklas9/projx-store => ../projx-store
