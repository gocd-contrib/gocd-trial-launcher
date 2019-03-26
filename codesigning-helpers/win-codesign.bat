cd %~dp0..\..\codesigning
md src\win
move %~dp0..\..\dist\windows\amd64\run-gocd.exe src\win
bundle install
bundle exec rake --trace win:sign
