$files = Get-ChildItem $env:TEMP -Filter "chromedp*"
foreach ($file in $files)
{
    Remove-Item $file.FullName -Recurse
    Write-Host "Deleted: " $file.FullName
}