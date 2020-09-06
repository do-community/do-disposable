# Make sure this is Windows.
if (-not (($host.name -eq "ConsoleHost") -or ($host.Name -eq "Windows PowerShell ISE Host"))) {
    throw "This script is designed for Windows only."
}

# Create the bin folder.
echo "Creating binary folder (a bit like /usr/local/bin on Linux) to allow binaries to easily be downloaded."
$binfolder = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $binfolder | out-null

# Add the bin folder to the PATH if not exists.
$path = (Get-ItemProperty -Path 'Registry::HKEY_CURRENT_USER\Environment' -Name PATH).path
if (-not ('$path' -Match '$binfolder')) {
    $path += ";$binfolder"
    Set-ItemProperty -Path 'Registry::HKEY_CURRENT_USER\Environment' -Name PATH -Value $path
}

# Get the do-disposable exe.
echo "Downloading do-disposable."
$processorType = "amd64"
if ($Env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $processorType = "arm"
} else {
    if ($Env:PROCESSOR_ARCHITECTURE -ne "AMD64") {
        $processorType = "386"
    }
}
$url = "https://github.com/do-community/do-disposable/releases/download/v1.0.0/do-disposable_windows-$processorType.exe"
Invoke-WebRequest -Uri $url -OutFile $binfolder\do-disposable.exe

# Echo done and wait for input.
echo "Install complete."
Write-Host -NoNewLine "Press any key to continue..."
cmd /c pause | out-null
