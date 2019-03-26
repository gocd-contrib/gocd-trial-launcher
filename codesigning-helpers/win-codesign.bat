cd %~dp0..\..\codesigning
md src\win
md out\win
move %~dp0..\..\dist\windows\amd64\run-gocd.exe src\win
gem install bundler
bundle install
bundle exec rake --trace win:sign
