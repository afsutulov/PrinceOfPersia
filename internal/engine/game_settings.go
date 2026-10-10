// SPDX-License-Identifier: GPL-3.0-or-later
package engine

import "modernc.org/libc"

// Original game settings and tables, expressed as Go data.
var originalGameSettings = custom_options_type{
	Fstart_minutes_left:         0x29a,
	Fstart_ticks_left:           0x2cf,
	Fstart_hitp:                 0x3,
	Fmax_hitp_allowed:           0xa,
	Fsaving_allowed_first_level: 0x1,
	Fsaving_allowed_last_level:  0xd,
	Fstart_upside_down:          0x0,
	Fstart_in_blind_mode:        0x0,
	Fcopyprot_level:             0xffff,
	Fdrawn_tile_top_level_edge:  0x1,
	Fdrawn_tile_left_level_edge: 0x14,
	Flevel_edge_hit_tile:        0x14,
	Fallow_triggering_any_tile:  0x0,
	Fenable_wda_in_palace:       0x0,
	Fvga_palette: [16]struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x0, Fg: 0x0, Fb: 0x0}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x0, Fg: 0x0, Fb: 0x2a}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x0, Fg: 0x2a, Fb: 0x0}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x0, Fg: 0x2a, Fb: 0x2a}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x2a, Fg: 0x0, Fb: 0x0}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x2a, Fg: 0x0, Fb: 0x2a}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x2a, Fg: 0x15, Fb: 0x0}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x2a, Fg: 0x2a, Fb: 0x2a}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x15, Fg: 0x15, Fb: 0x15}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x15, Fg: 0x15, Fb: 0x3f}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x15, Fg: 0x3f, Fb: 0x15}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x15, Fg: 0x3f, Fb: 0x3f}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x3f, Fg: 0x15, Fb: 0x15}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x3f, Fg: 0x15, Fb: 0x3f}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x3f, Fg: 0x3f, Fb: 0x15}, struct {
		Fr uint8
		Fg uint8
		Fb uint8
	}{Fr: 0x3f, Fg: 0x3f, Fb: 0x3f}},
	Ffirst_level:                      0x1,
	Fskip_title:                       0x0,
	Fshift_L_allowed_until_level:      0x4,
	Fshift_L_reduced_minutes:          0xf,
	Fshift_L_reduced_ticks:            0x2cf,
	Fdemo_hitp:                        0x4,
	Fdemo_end_room:                    0x18,
	Fintro_music_level:                0x1,
	Fhave_sword_from_level:            0x2,
	Fcheckpoint_level:                 0x3,
	Fcheckpoint_respawn_dir:           -1,
	Fcheckpoint_respawn_room:          0x2,
	Fcheckpoint_respawn_tilepos:       0x6,
	Fcheckpoint_clear_tile_room:       0x7,
	Fcheckpoint_clear_tile_col:        0x4,
	Fcheckpoint_clear_tile_row:        0x0,
	Fskeleton_level:                   0x3,
	Fskeleton_room:                    0x1,
	Fskeleton_trigger_column_1:        0x2,
	Fskeleton_trigger_column_2:        0x3,
	Fskeleton_column:                  0x5,
	Fskeleton_row:                     0x1,
	Fskeleton_require_open_level_door: 0x1,
	Fskeleton_skill:                   0x2,
	Fskeleton_reappear_room:           0x3,
	Fskeleton_reappear_x:              0x85,
	Fskeleton_reappear_row:            0x1,
	Fskeleton_reappear_dir:            0x0,
	Fmirror_level:                     0x4,
	Fmirror_room:                      0x4,
	Fmirror_column:                    0x4,
	Fmirror_row:                       0x0,
	Fmirror_tile:                      0xd,
	Fshow_mirror_image:                0x1,
	Fshadow_steal_level:               0x5,
	Fshadow_steal_room:                0x18,
	Fshadow_step_level:                0x6,
	Fshadow_step_room:                 0x1,
	Ffalling_exit_level:               0x6,
	Ffalling_exit_room:                0x1,
	Ffalling_entry_level:              0x7,
	Ffalling_entry_room:               0x11,
	Fmouse_level:                      0x8,
	Fmouse_room:                       0x10,
	Fmouse_delay:                      0x96,
	Fmouse_object:                     0x18,
	Fmouse_start_x:                    0xc8,
	Floose_tiles_level:                0xd,
	Floose_tiles_room_1:               0x17,
	Floose_tiles_room_2:               0x10,
	Floose_tiles_first_tile:           0x16,
	Floose_tiles_last_tile:            0x1b,
	Fjaffar_victory_level:             0xd,
	Fjaffar_victory_flash_time:        0x12,
	Fhide_level_number_from_level:     0xe,
	Flevel_13_level_number:            0xc,
	Fvictory_stops_time_level:         0xd,
	Fwin_level:                        0xe,
	Fwin_room:                         0x5,
	Floose_floor_delay:                0xb,
	Ftbl_level_type:                   [16]uint8{0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x1, 0x0},
	Ftbl_level_color:                  [16]uint16{0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x2, 0x2, 0x0, 0x0, 0x3, 0x3, 0x4, 0x0},
	Ftbl_guard_type:                   [16]int16{0, 0, 0, 2, 0, 0, 1, 0, 0, 0, 0, 0, 4, 3, -1, -1},
	Ftbl_guard_hp:                     [16]uint8{0x4, 0x3, 0x3, 0x3, 0x3, 0x4, 0x5, 0x4, 0x4, 0x5, 0x5, 0x5, 0x4, 0x6, 0x0, 0x0},
	Ftbl_cutscenes_by_index:           [16]uint8{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	Ftbl_entry_pose:                   [16]uint8{0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0},
	Ftbl_seamless_exit:                [16]int8{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 23, -1, -1, -1},
	Fstrikeprob:                       [12]uint16{0x4b, 0x64, 0x4b, 0x4b, 0x4b, 0x32, 0x64, 0xdc, 0x0, 0x3c, 0x28, 0x3c},
	Frestrikeprob:                     [12]uint16{0x0, 0x0, 0x0, 0x5, 0x5, 0xaf, 0x14, 0xa, 0x0, 0xff, 0xff, 0x96},
	Fblockprob:                        [12]uint16{0x0, 0x96, 0x96, 0xc8, 0xc8, 0xff, 0xc8, 0xfa, 0x0, 0xff, 0xff, 0xff},
	Fimpblockprob:                     [12]uint16{0x0, 0x4b, 0x4b, 0x64, 0x64, 0x91, 0x64, 0xfa, 0x0, 0x91, 0xff, 0xaf},
	Fadvprob:                          [12]uint16{0xff, 0xc8, 0xc8, 0xc8, 0xff, 0xff, 0xc8, 0x0, 0x0, 0xff, 0x64, 0x64},
	Frefractimer:                      [12]uint16{0x14, 0x14, 0x14, 0x14, 0xa, 0xa, 0xa, 0xa, 0x0, 0xa, 0x0, 0x0},
	Fextrastrength:                    [12]uint16{0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	Finit_shad_6:                      [8]uint8{0xf, 0x51, 0x76, 0x0, 0x0, 0x1, 0x0, 0x0},
	Finit_shad_5:                      [8]uint8{0xf, 0x37, 0x37, 0x0, 0xff, 0x0, 0x0, 0x0},
	Finit_shad_12:                     [8]uint8{0xf, 0x51, 0xe8, 0x0, 0x0, 0x0, 0x0, 0x0},
	Fdemo_moves: [25]struct {
		Ftime int16
		Fmove int16
	}{struct {
		Ftime int16
		Fmove int16
	}{Ftime: 0, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 1, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 13, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 30, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 37, Fmove: 5}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 47, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 48, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 65, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 73, Fmove: 2}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 75, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 99, Fmove: 2}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 100, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 115, Fmove: 5}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 128, Fmove: 6}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 136, Fmove: 3}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 157, Fmove: 7}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 158, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 159, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 171, Fmove: 4}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 177, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 178, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 188, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 193, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 205, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 233, Fmove: -1}},
	Fshad_drink_move: [8]struct {
		Ftime int16
		Fmove int16
	}{struct {
		Ftime int16
		Fmove int16
	}{Ftime: 0, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 1, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 14, Fmove: 0}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 18, Fmove: 6}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 29, Fmove: 7}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 45, Fmove: 2}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 49, Fmove: 1}, struct {
		Ftime int16
		Fmove int16
	}{Ftime: 255, Fmove: -2}},
	Fbase_speed:         0x5,
	Ffight_speed:        0x6,
	Fchomper_speed:      0xf,
	Fno_mouse_in_ending: 0x0,
}

func applyGameSettings(tls *libc.TLS) {
	custom_saved = originalGameSettings
	turn_custom_options_on_off(tls, 1)
}
