# GoTTY releases

## v1.3.0

 * Links in the tty are now clickable.
 * Use WebGL for rendering by default.
 * Ensure authentication (TLS or Basic auth) remain enabled even if some of the options are only given in config files Thanks, @devanlai!
 * Fix typo in README.md Thanks, @prusnak!
 * Add arm64/Linux build. Thanks for the suggestion, @nephaste!

## v1.2.0

 * Pass BUILD\_OPTIONS to gox, too, so release artifacts have version info included.
 * Update xterm.js 2.7.0 => 4.11.0
 * Lots of clean up.

## v1.1.0

 * Today I learned about Go's handling of versions, so re-releasing 2.1.0 as 1.1.0.
 * Added path option. Thanks, @apatil!

## v2.1.0 (whoops)

 * Use Go modules and update cli module import path. Thanks, @svanellewee!
 * Fix typos. Thanks, @0xflotus, @RealCyGuy, @ygit, @Jason-Cooke and @fredster33!
 * Fix printing of ipv6 addresses. Thanks, @Felixoid!
 * Add Progressive Web App support. Thanks, @sehaas!
 * Add instructions for GNU screen. Thanks, @Immortalin!
 * Add Solaris support. Thanks, @fazalmajid!
 * New maintainer: @sorenisanerd
