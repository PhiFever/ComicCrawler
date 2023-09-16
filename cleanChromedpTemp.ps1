$files = Get-ChildItem $env:TEMP -Filter "chromedp*"
foreach ($file in $files)
{
    Remove-Item $file.FullName -Recurse
}