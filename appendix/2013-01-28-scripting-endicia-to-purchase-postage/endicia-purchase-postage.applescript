on run argv
    tell application "System Events"
        if UI elements enabled is false then
            tell application "System Preferences"
                activate

                set current pane to pane id "com.apple.preference.universalaccess"

                display dialog "This feature cannot be used until \"Enable access for assistive devices\" is enabled here. Please make that change and then try again." with icon 1 buttons {"OK"} default button "OK"

                return "error"
            end tell
        end if
    end tell

    tell application "Endicia"
        activate

        tell application "System Events" to tell application process "Endicia"
            tell menu bar 1
                tell menu bar item "Postage"
                    pick
                    tell menu 1
                        pick menu item "Buy Postage…"
                    end tell
                end tell
            end tell

            delay 1

            tell window "Buy Postage"
                tell pop up button 1
                    click
                    tell menu 1
                        click menu item ("$" & (item 1 of argv) & ".00")
                    end tell
                end tell

                tell button "Purchase Postage Now"
                    click
                end tell
            end tell

            delay 1

            tell window 1
                tell button "OK"
                    click
                end tell
            end tell

            return "ok"
        end tell
    end tell
end run
