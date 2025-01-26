:: setup
cd %~dp0..\..\codesigning
call rake --trace win:metadata_single_binary[../dist/windows/amd64/run-gocd.exe,../win-launcher.zip]
