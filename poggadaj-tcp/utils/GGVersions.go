package utils

import "poggadaj-tcp/constants"

//var GGVersions = map[uint32]string{
//	1073741857: "Gadu-Gadu 6.0 (build 139)",
//	1073741860: "Gadu-Gadu 6.1 (build 158)",
//	1073741863: "Gadu-Gadu 7.0/7.1",
//	3254779944: "Gadu-Gadu 7.5 (build 2201)",
//	1073872938: "Gadu-Gadu 7.7 (build 3746)",
//}

var GGVersions = map[uint8]string{
	0x0b: "Gadu-Gadu 4.0.25 - 4.0.30",
	0x0f: "Gadu-Gadu 4.5.12",
	0x10: "Gadu-Gadu 4.5.15 - 4.5.22",
	0x11: "Gadu-Gadu 4.6.1 - 4.6.10",
	0x14: "Gadu-Gadu 4.8.3 - 4.8.1",
	0x15: "Gadu-Gadu 4.8.9",
	0x16: "Gadu-Gadu 4.9.1",
	0x17: "Gadu-Gadu 4.9.2",
	0x18: "Gadu-Gadu 4.9.3 - 5.0.1",
	0x19: "Gadu-Gadu 5.0.3",
	0x1b: "Gadu-Gadu 5.0.5",
	0x1c: "Gadu-Gadu 5.7 beta",
	0x1e: "Gadu-Gadu 5.7 beta (build 121)",
	0x20: "Gadu-Gadu 6.0",
	0x21: "Gadu-Gadu 6.0 (build 133)",
	0x22: "Gadu-Gadu 6.0 (build 140)",
	0x24: "Gadu-Gadu 6.1 (build 155)",
	0x25: "Gadu-Gadu 7.0 (build 1)",
	0x26: "Gadu-Gadu 7.0 (build 20)",
	0x27: "Gadu-Gadu 7.0 (build 22)",
	0x28: "Gadu-Gadu 7.5 (build 2201)",
	0x29: "Gadu-Gadu 7.6 (build 1688)",
	0x2a: "Gadu-Gadu 7.7 (build 3315)",
	0x2d: "Gadu-Gadu 8.0 (build 4881)",
}

func GetVersionAndVOIP(version uint32) (uint8, bool) {
	if constants.GG_HAS_AUDIO_MASK&version == constants.GG_HAS_AUDIO_MASK {
		return uint8(version), true
	}
	return uint8(version), false
}
