:: setup
cd %~dp0..\..\codesigning
md src\win
move %~dp0..\..\dist\windows\amd64\run-gocd.exe src\win
call gem install bundler
call bundle install
call bundle exec rake --trace win:sign
:: package
cd out\win
call jar -cMf ..\..\..\win-launcher.zip run-gocd.exe
