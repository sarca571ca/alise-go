### TODO

- [x] Implement camp channel commands.
- [x] Consolidate the ephemeral response to one single function eg. respondError() and respondWithLinkshellList() are practically the same.
- [x] Implement logging, with commad usage tracking.
- [ ] Expand the logging to cover other sections of the bot using a wrapper. Like the one for commands.
- [x] Remove the InteractionResponse on the HNM command we don't need the current timers listed in the bot commands channel. Maybe a ephemeral confirmation instead?
- [ ] Fix the formating on the Linkshell List. Maybe implement a leaderboard channel instead.
- [x] Turn the output of /linkshell list to just give the linkshells currently tracked.
- [ ] When implementing the claim leaderboard can show the claims for the week, month and all time.
- [ ] Make the HNM Board look better by adding pictures etc.
- [ ] Need to make checks when dealing with the Linkshell tables in the db for capitalization etc. to prevent duplicates.
- [ ] Updating timers within the 20 minutes before camp will cause a 2nd channel to be created. Need to check if a channel is already made and use that one.
- [ ] If the bot goes down need to reschedule any channels marked MoveScheduled in the HNMTimesCategory to be moved.
- [ ] Manage grand wyrvn windows properly. Look at EnrageWindow usage from the enrage command.
- [ ] Fix camp channels to be nicer
- [ ] Timers dont seem to sort in order correctly on the hnm timer board
