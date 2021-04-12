const { spawn } = require('child_process')
const { platform, arch } = require('os')
const { join } = require('path')
const hostos = platform() === 'win32' ? 'windows' : platform()
const archMapping = {
	x64: 'amd64',
	x32: '386',
	ia32: '386',
}
const hostarch = archMapping[arch()] !== undefined ? archMapping[arch()] : arch()
const ext = hostos === 'windows' ? '.exe' : ''
const args = ['azure', 'start']
const stucco = spawn(join('stucco', hostos, hostarch, 'stucco' + ext), args.concat(process.argv.slice(2)))
process.stdin.pipe(stucco.stdin)
stucco.stdout.pipe(process.stdout)
stucco.stderr.pipe(process.stderr)
