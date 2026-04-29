### TODO

- [ ] Implement camp channel commands.
- [x] Consolidate the ephemeral response to one single function eg. respondError() and respondWithLinkshellList() are practically the same.
- [x] Implement logging, with commad usage tracking.
- [ ] Expand the logging to cover other sections of the bot using a wrapper. Like the one for commands.
- [x] Remove the InteractionResponse on the HNM command we don't need the current timers listed in the bot commands channel. Maybe a ephemeral confirmation instead?
- [ ] Fix the formating on the Linkshell List. Maybe implement a leaderboard channel instead.
- [ ] Turn the output of /linkshell list to just give the linkshells currently tracked.
- [ ] When implementing the claim leaderboard can show the claims for the week, month and all time.
- [ ] Make the HNM Board look better by adding pictures etc.
- [ ] Need to make checks when dealing with the Linkshell tables in the db for capitalization etc. to prevent duplicates.
