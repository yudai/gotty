## 1.4.0 (2022-05-30)

*  Add @hardliner66 as a contributor ([1ca998e](https://github.com/sorenisanerd/gotty/commit/1ca998e))
*  Add @jkandasa as a contributor ([cd23910](https://github.com/sorenisanerd/gotty/commit/cd23910))
* Add backend tests ([603c650](https://github.com/sorenisanerd/gotty/commit/603c650))
* Add generated data to git ([a9fbc07](https://github.com/sorenisanerd/gotty/commit/a9fbc07))
* add quiet flag to disable logging ([4109b11](https://github.com/sorenisanerd/gotty/commit/4109b11))
* Add references to @yudai ([bffd821](https://github.com/sorenisanerd/gotty/commit/bffd821)), closes [#8](https://github.com/sorenisanerd/gotty/issues/8)
* Add rule to build gotty.js.map ([82c3acf](https://github.com/sorenisanerd/gotty/commit/82c3acf))
* Apply font size and family in xterm ([f157dbe](https://github.com/sorenisanerd/gotty/commit/f157dbe)), closes [#21](https://github.com/sorenisanerd/gotty/issues/21)
* Avoid HTTP 401 error on manifest.json due to CORS ([817b5c8](https://github.com/sorenisanerd/gotty/commit/817b5c8))
* Bump browserslist from 4.16.4 to 4.16.6 in /js ([8deba62](https://github.com/sorenisanerd/gotty/commit/8deba62))
* Disable arg passing by default ([5c8eb10](https://github.com/sorenisanerd/gotty/commit/5c8eb10)), closes [#17](https://github.com/sorenisanerd/gotty/issues/17)
* Do not include ALL of bootstrap ([b63ea16](https://github.com/sorenisanerd/gotty/commit/b63ea16))
* Ensure defaults for booleans is set correctly ([28f8e61](https://github.com/sorenisanerd/gotty/commit/28f8e61)), closes [#16](https://github.com/sorenisanerd/gotty/issues/16)
* Fix existing tests ([d674aa1](https://github.com/sorenisanerd/gotty/commit/d674aa1)), closes [#13](https://github.com/sorenisanerd/gotty/issues/13)
* Fix warnings from Markdown linter ([aa86a34](https://github.com/sorenisanerd/gotty/commit/aa86a34))
* go fmt ([dcb153c](https://github.com/sorenisanerd/gotty/commit/dcb153c))
* Improve webtty test coverage ([f61763f](https://github.com/sorenisanerd/gotty/commit/f61763f))
* Make client request base64 encoding ([dd3603c](https://github.com/sorenisanerd/gotty/commit/dd3603c))
* Make sure we read the full message ([1eed97f](https://github.com/sorenisanerd/gotty/commit/1eed97f))
* Publish artifacts on push to master ([6c62ab7](https://github.com/sorenisanerd/gotty/commit/6c62ab7))
* Remove hterm ([163fd05](https://github.com/sorenisanerd/gotty/commit/163fd05))
* Run tests on push ([55674f1](https://github.com/sorenisanerd/gotty/commit/55674f1))
* Run tests on push to all branches ([679a324](https://github.com/sorenisanerd/gotty/commit/679a324))
* update go version in Dockerfile ([fd2fb99](https://github.com/sorenisanerd/gotty/commit/fd2fb99))
* Update js dependencies ([26fc412](https://github.com/sorenisanerd/gotty/commit/26fc412))
* Update xterm.js and other js libs ([81afdc7](https://github.com/sorenisanerd/gotty/commit/81afdc7)), closes [#18](https://github.com/sorenisanerd/gotty/issues/18)
* Use bootstrap components for up- and downloads ([7f05f2f](https://github.com/sorenisanerd/gotty/commit/7f05f2f))
* Use Go's built-in embed mechanism ([f66f0d0](https://github.com/sorenisanerd/gotty/commit/f66f0d0)), closes [#7](https://github.com/sorenisanerd/gotty/issues/7)
* feat(zmodem): Allow file uploads/downloads ([782991c](https://github.com/sorenisanerd/gotty/commit/782991c))



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
