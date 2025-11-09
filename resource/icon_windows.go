//go:build windows

package resource

/*
#include <windows.h>
void set_resource_icon(const void *ptr, char* name) {
	HINSTANCE hInst = GetModuleHandle(NULL);
	HICON iconBig = (HICON)LoadImage(hInst, name, IMAGE_ICON, GetSystemMetrics(SM_CXICON), GetSystemMetrics(SM_CXICON), LR_DEFAULTCOLOR);
	HICON iconSml = (HICON)LoadImage(hInst, name, IMAGE_ICON, GetSystemMetrics(SM_CXSMICON), GetSystemMetrics(SM_CYSMICON), LR_DEFAULTCOLOR);
	if (iconSml) SendMessage((HWND)ptr, WM_SETICON, ICON_SMALL, (LPARAM)iconSml);
	if (iconBig) SendMessage((HWND)ptr, WM_SETICON, ICON_BIG, (LPARAM)iconBig);
}
*/
import "C"
import (
	"unsafe"

	webview "github.com/webview/webview_go"
)

func SetAppIcon(w webview.WebView) {
	hwnd := w.Window()
	cstr := C.CString("#1")
	defer C.free(unsafe.Pointer(cstr))
	C.set_resource_icon(hwnd, cstr)
}
