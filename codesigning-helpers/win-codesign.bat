:: setup
cd %~dp0..\..\codesigning
md src\win
move %~dp0..\..\dist\windows\amd64\run-gocd.exe src\win
gem install bundler
bundle install

bundle exec rake --trace win:sign

:: package
cd out\win
jar -cMf ..\..\..\win-launcher.zip run-gocd.exe
