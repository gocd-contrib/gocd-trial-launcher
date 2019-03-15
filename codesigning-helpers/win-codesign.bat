cd %~dp0\..\..\codesigning
md src\win
move %~dp0\..\..\dist\windows\amd64\run-gocd.exe src\win
rake --trace win:sign
dir out\win
cd out\win
jar --help
jar -cMf %~dp0\..\..\win.zip run-gocd.exe -C out\win
