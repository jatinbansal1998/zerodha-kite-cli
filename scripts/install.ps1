[CmdletBinding()]
param(
  [string]$Version = $env:ZERODHA_VERSION,
  [string]$InstallDir = $env:ZERODHA_INSTALL_DIR,
  [string]$Repo = $env:ZERODHA_REPO
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = 'latest'
}

if ([string]::IsNullOrWhiteSpace($Repo)) {
  $Repo = 'jatinbansal1998/zerodha-kite-cli'
}

if ([string]::IsNullOrWhiteSpace($InstallDir)) {
  $InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\zerodha-kite-cli\bin'
}

$archRaw = $env:PROCESSOR_ARCHITECTURE
if (-not [string]::IsNullOrWhiteSpace($env:PROCESSOR_ARCHITEW6432)) {
  $archRaw = $env:PROCESSOR_ARCHITEW6432
}

$targetArch = switch ($archRaw.ToUpperInvariant()) {
  'AMD64' { 'amd64' }
  'X86_64' { 'amd64' }
  'ARM64' { 'arm64' }
  default { throw "Unsupported CPU architecture: $archRaw" }
}

$assetName = "zerodha_windows_${targetArch}.exe"

if ($Version -eq 'latest') {
  $downloadUrl = "https://github.com/$Repo/releases/latest/download/$assetName"
} else {
  if (-not $Version.StartsWith('v')) {
    $Version = "v$Version"
  }
  $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$assetName"
}

New-Item -Path $InstallDir -ItemType Directory -Force | Out-Null
$tempFile = Join-Path ([System.IO.Path]::GetTempPath()) ("zerodha-install-" + [System.Guid]::NewGuid().ToString('N') + ".exe")
$destFile = Join-Path $InstallDir 'zerodha.exe'

Write-Host "Downloading $downloadUrl"
try {
  Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
  Move-Item -Path $tempFile -Destination $destFile -Force
} finally {
  if (Test-Path $tempFile) {
    Remove-Item -Path $tempFile -Force
  }
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$pathUpdated = $false

$installFull = [System.IO.Path]::GetFullPath($InstallDir).TrimEnd('\')
$pathContainsInstall = $false

if (-not [string]::IsNullOrWhiteSpace($userPath)) {
  $segments = $userPath -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
  foreach ($segment in $segments) {
    try {
      $segmentFull = [System.IO.Path]::GetFullPath($segment).TrimEnd('\')
      if ($segmentFull -ieq $installFull) {
        $pathContainsInstall = $true
        break
      }
    } catch {
      continue
    }
  }
}

if (-not $pathContainsInstall) {
  if ([string]::IsNullOrWhiteSpace($userPath)) {
    $newPath = $InstallDir
  } else {
    $newPath = "$userPath;$InstallDir"
  }
  [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
  $pathUpdated = $true
}

Write-Host "Installed zerodha to $destFile"
if ($pathUpdated) {
  Write-Host "Added $InstallDir to user PATH. Restart the terminal to use 'zerodha'."
} else {
  Write-Host "Run: zerodha version"
}
