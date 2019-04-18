:: setup
cd %~dp0..\..\codesigning
call gem install bundler
call bundle install
call bundle exec rake --trace win:sign_single_binary[..\..\dist\windows\amd64\run-gocd.exe,..\..\..\win-launcher.zip]
