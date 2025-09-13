$ErrorActionPreference = 'Stop'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

Remove-Item "$toolsDir\hottgo.exe" -Force -ErrorAction SilentlyContinue
