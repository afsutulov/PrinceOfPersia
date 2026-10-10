// SPDX-License-Identifier: GPL-3.0-or-later
// Exact unsigned 32-bit update from SDLPoP seg009.c:prandom.
package dosrng

func Next(seed uint32, max uint16) (uint32, uint16) {
	seed = seed*214013 + 2531011
	return seed, uint16((seed >> 16) % (uint32(max) + 1))
}
