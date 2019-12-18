# About

GraphQL server runner created with serverless in mind. Project is in early alpha phase. Backwards compatibility is not guaranteed.

# Installation

* macOS/Linux
	```
	$ curl https://stucco-release.fra1.cdn.digitaloceanspaces.com/latest/$(uname | tr '[:upper:]' '[:lower:]')/$(uname -m | sed 's/^x86_64$/amd64/g)/stucco
	```

* Windows
	[Download for 64-bit](https://stucco-release.fra1.cdn.digitaloceanspaces.com/latest/windows/amd64/stucco.exe)
# Drivers

To add a new provider/runtime implement `driver.Driver`

# Example

An example project using `stucco-js` driver is available in example.
