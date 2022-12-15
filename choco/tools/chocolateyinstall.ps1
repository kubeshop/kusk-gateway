$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64      = "https://github.com/kubeshop/kusk-gateway/releases/download/v${env:chocolateyPackageVersion}/kusk_${env:chocolateyPackageVersion}_Windows_x86_64.tar.gz" # 64bit URL here (HTTPS preferred) or remove - if installer contains both (very rare), use $url

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  fileType      = 'exe'
  url64bit      = $url64
  softwareName  = 'kusk*'

  # %placeholder% will be replaced by the correct value in the CI pipeline
  checksum      = '%checksum%'
  checksumType  = 'sha256'

}

Install-ChocolateyZipPackage @packageArgs
Get-ChocolateyUnzip -FileFullPath "$toolsDir/kusk_${env:chocolateyPackageVersion}_Windows_x86_64.tar" -Destination $toolsDir
