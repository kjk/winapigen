// THIS FILE WAS AUTO-GENERATED BY https://github.com/kjk/w/cmd/gengo
package w

const (
	CP_ACP              = 0
	CP_OEMCP            = 1
	CP_MACCP            = 2
	CP_THREAD_ACP       = 3
	CP_SYMBOL           = 42
	MS_DOS_Latin_US     = 437
	Thai                = 874
	Japanese_Shift_JIS  = 932
	Chinese_Simplified  = 936
	Korean              = 949
	Chinese_Traditional = 950
	Unicode_UTF_16_LE   = 1200
	Unicode_UTF_16_BE   = 1201
	Central_European    = 1250
	Cyrillic            = 1251
	Western_European    = 1252
	Greek               = 1253
	Turkish             = 1254
	Hebrew              = 1255
	Arabic              = 1256
	Baltic              = 1257
	Vietnamese          = 1258
	CP_UTF7             = 65000
	CP_UTF8             = 65001
)

type HWND HANDLE

const (
	NULL = 0
)

const (
	OWNER_SECURITY_INFORMATION            = 0x00000001
	GROUP_SECURITY_INFORMATION            = 0x00000002
	DACL_SECURITY_INFORMATION             = 0x00000004
	SACL_SECURITY_INFORMATION             = 0x00000008
	LABEL_SECURITY_INFORMATION            = 0x00000010
	PROTECTED_DACL_SECURITY_INFORMATION   = 0x80000000
	PROTECTED_SACL_SECURITY_INFORMATION   = 0x40000000
	UNPROTECTED_DACL_SECURITY_INFORMATION = 0x20000000
	UNPROTECTED_SACL_SECURITY_INFORMATION = 0x10000000
)

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

type SYSTEMTIME struct {
	WYear         uint16
	WMonth        uint16
	WDayOfWeek    uint16
	WDay          uint16
	WHour         uint16
	WMinute       uint16
	WSecond       uint16
	WMilliseconds uint16
}

type TIME_ZONE_INFORMATION struct {
	Bias         int32
	StandardName [32]WCHAR
	StandardDate SYSTEMTIME
	StandardBias int32
	DaylightName [32]WCHAR
	DaylightDate SYSTEMTIME
	DaylightBias int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

const (
	GetFileExInfoStandard = 0
)

const (
	FILE_ATTRIBUTE_READONLY            = 0x00000001
	FILE_ATTRIBUTE_HIDDEN              = 0x00000002
	FILE_ATTRIBUTE_SYSTEM              = 0x00000004
	FILE_ATTRIBUTE_DIRECTORY           = 0x00000010
	FILE_ATTRIBUTE_ARCHIVE             = 0x00000020
	FILE_ATTRIBUTE_DEVICE              = 0x00000040
	FILE_ATTRIBUTE_NORMAL              = 0x00000080
	FILE_ATTRIBUTE_TEMPORARY           = 0x00000100
	FILE_ATTRIBUTE_SPARSE_FILE         = 0x00000200
	FILE_ATTRIBUTE_REPARSE_POINT       = 0x00000400
	FILE_ATTRIBUTE_COMPRESSED          = 0x00000800
	FILE_ATTRIBUTE_OFFLINE             = 0x00001000
	FILE_ATTRIBUTE_NOT_CONTENT_INDEXED = 0x00002000
	FILE_ATTRIBUTE_ENCRYPTED           = 0x00004000
	FILE_ATTRIBUTE_VIRTUAL             = 0x00010000
	FILE_ATTRIBUTE_NO_SCRUB_DATA       = 0x00020000
	INVALID_FILE_ATTRIBUTES            = -1
)

const (
	IO_REPARSE_TAG_MOUNT_POINT = 0xA0000003
	IO_REPARSE_TAG_HSM         = 0xC0000004
	IO_REPARSE_TAG_HSM2        = 0x80000006
	IO_REPARSE_TAG_SIS         = 0x80000007
	IO_REPARSE_TAG_WIM         = 0x80000008
	IO_REPARSE_TAG_CSV         = 0x80000009
	IO_REPARSE_TAG_DFS         = 0x8000000A
	IO_REPARSE_TAG_SYMLINK     = 0xA000000C
	IO_REPARSE_TAG_DFSR        = 0x80000012
	IO_REPARSE_TAG_DEDUP       = 0x80000013
	IO_REPARSE_TAG_NFS         = 0x80000014
)

type WIN32_FIND_DATAW struct {
	DwFileAttributes   uint32
	FtCreationTime     FILETIME
	FtLastAccessTime   FILETIME
	FtLastWriteTime    FILETIME
	NFileSizeHigh      uint32
	NFileSizeLow       uint32
	DwReserved0        uint32
	DwReserved1        uint32
	CFileName          [260]WCHAR
	CAlternateFileName [14]WCHAR
}
