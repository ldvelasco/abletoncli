# Repository instructions

## Commands

- Build the Go CLI: `go build -o abletoncli.exe .`
- Run Go tests: `go test ./...`; focus a test with `go test -run TestName .`.
- AbletonOSC tests are integration tests: from `abletonosc/`, run `pytest` with Live open, AbletonOSC selected, default audio devices configured, and Count-In set to `None`.

## Structure and integration

- The root is a dependency-free Go CLI (`main.go`, command translation in `commands.go`, OSC wire/client code in `osc.go`). `abletonosc/` is the Python Remote Script that implements the OSC API; CLI features may need matching changes there.
- OSC queries bind response port `11001`; run queries serially, not concurrently.
- To install the modified Remote Script, copy the *contents* of `abletonosc/` to Live's `User Library/Remote Scripts/AbletonOSC` directory and restart Live. Copying only its inner `abletonosc/` package is insufficient.
- For Arrangement editing workflows and Live-version quirks, consult `skills/ableton-arrangement/SKILL.md`.
