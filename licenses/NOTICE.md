# Credits and third-party notices

PrinceGo's recovered gameplay engine is based on SDLPoP by Dávid Nagy and contributors, revision `3c5add5fb7f83d4ceb542823ab66d00146c4271b`:
https://github.com/NagyD/SDLPoP
The game source is GPL-3.0-or-later. Original copyright notices remain in the translated Go source. The translation was produced with modernc.org/ccgo/v4 v4.36.1; ccgo is not needed to build this project.

Ebitengine v2.10.0 supplies desktop graphics, input and audio interfaces (MIT). Oto v3.5.0 and purego v0.11.0 use Apache-2.0. modernc.org/libc v1.76.0 provides Go runtime compatibility. Exact dependency versions are in go.mod/go.sum; upstream licenses and notices are in this folder and dependencies/.

Nuked OPL3 by Alexey Khokholov is GPL-2.0-or-later. stb_vorbis by Sean Barrett and contributors is MIT/public domain. SDLPoP MIDI integration also credits Chocolate Doom and Falcury. The translated declarations retain glibc header notices (LGPL-2.1-or-later). The Go raster adapter references SDL2 surface semantics; SDL2 uses the zlib license. No SDL library is distributed or loaded.

The optional Windows resource generation tool rsrc v0.10.2 is MIT, Copyright 2013–2020 The rsrc Authors:
https://github.com/akavel/rsrc
It runs only while generating a Windows icon resource and is not a runtime dependency.

The original Prince of Persia resources were supplied by the user. Prince of Persia: Jordan Mechner / Brøderbund. Graphics, levels and sound retain their original rights; GPL licensing of the engine does not grant a separate asset license. No DOS executable or DOS executable machine code is included in the embedded resource collection. Game settings are represented as Go data.

The PrinceGo application icon was generated for this project. Screenshots show the running Go game. The project is not endorsed by the original game authors or SDLPoP maintainers.
