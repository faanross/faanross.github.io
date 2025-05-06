---
showTableOfContents: true
title: "Introducing Basic Delays and Misdirection (Theory 9.2)"
type: "page"
---
## Introducing Delays/Misdirection

So while separating the Write (RW) and Execute (RX) phases by using `VirtualProtect` is definitely a step up from the blatant RWX allocation, it's still not exactly nation-state stealth we're employing here. Think about it from an EDR's perspective. It might have hooks or monitoring points on functions like `VirtualAlloc` and `VirtualProtect`. Seeing a call sequence like:

1. `VirtualAlloc` (requesting `PAGE_READWRITE` for region X)
2. `memcpy` / `RtlCopyMemory` (writing data into region X)
3. `VirtualProtect` (changing region X to `PAGE_EXECUTE_READ`)
4. An attempt to execute code from region X (e.g., via function pointer, `CreateThread`, etc.)

...happening in quick succession on the _exact same memory region_ is still a strong behavioural indicator. It clearly signals the intent to load and execute new code dynamically.

If the _sequence_ is suspicious, how can we make it less obvious? The goal is to break the clear, immediate chain of events that points directly from allocation to execution preparation. We want to introduce "noise" or "distance" between these actions, hoping to make automated correlation harder. Let's explore a handful of these actions to give you some idea of the general strategy.


## Time Delays (`Sleep`)
The simplest approach is to introduce a pause between writing the shellcode and changing the memory protection.


```cpp
// ... after RtlCopyMemory ...
    
Sleep(2000); // Pause for 2000 milliseconds (2 seconds)
    
// Now, change protection
if (!VirtualProtect(exec_memory, sizeof(calc_shellcode), PAGE_EXECUTE_READ, &oldProtect)) {
    // ... error handling ...
}
    
// ... execute ...
```

The question then is: how long to sleep? Too short, and it has no effect. Too long, and it might look suspicious itself (why is this thread randomly sleeping?) or hinder the implant's operation.

The overall idea here is to break the temporal link - an EDR logging API calls might see the `VirtualAlloc`, then the `memcpy`, then a gap, then `VirtualProtect`. This _might_ make simple time-based correlation less effective, but don't bet your farm on it - EDRs don't just rely on time; they correlate based on the target memory handle/address. So, even with a delay, the actions are still linked to the same region. In fact, `Sleep` itself can sometimes be monitored or considered an indicator in certain contexts, so you should def test it against the target EDR in a controlled environment to see what, if any, effect it would have.


## Dummy API Calls (Misdirection)

Instead of just sleeping, perform some seemingly innocuous actions between the write and the `VirtualProtect`.

```cpp
// ... after RtlCopyMemory ...
    
// --- Start Dummy Calls ---
DWORD tickCount = GetTickCount(); // Get system uptime
SYSTEMTIME sysTime;
GetSystemTime(&sysTime); // Get current UTC time
    
// Maybe perform a check that doesn't actually change state
// For example, query a common, non-sensitive registry key
HKEY hKey;
if (RegOpenKeyEx(HKEY_LOCAL_MACHINE, L"SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", 0, KEY_READ, &hKey) == ERROR_SUCCESS) {
    RegCloseKey(hKey);
}
// --- End Dummy Calls ---
    
    
// Now, change protection
if (!VirtualProtect(exec_memory, sizeof(calc_shellcode), PAGE_EXECUTE_READ, &oldProtect)) {
    // ... error handling ...
}
    
// ... execute ...
```


The basic idea here is just to "do stuff" before changing memory permissions. Not obviously this "stuff" should not impact our ability to execute our intended logic, and it should not be suspicious in and of itself. So choosing _good_ dummy calls is key. They should be common, low-impact APIs, calling obscure or privileged APIs would be counterproductive.

You should be aware that doing so can of course add some minor resource overhead, and that, as was the case above, correlation based on the specific target memory region can bypass this type of misdirection.


## Logical Decoupling (More Complex)

The basic idea here is that you structure your code so the allocation happens much earlier or in a different logical block than the eventual write, protect, and execute sequence. For example, memory might be allocated during an initialization phase, but the shellcode is only written and executed much later, perhaps triggered by a specific command from your C2 server. This creates significant logical and temporal distance, making correlation much harder but increases code complexity.


## Conclusion

Though, as I've alluded to, these techniques come with many "what-if" strings attached, I do think it's still useful to implement it as some of the meta-instruction can help provide insight onto "ways of thinking about solving problems" as we move ahead into more advanced territories. Let's jump in a quick lab to decouple permissions and add delays + misdirection.

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "decouple.md" >}})
[|NEXT|]({{< ref "decouple_lab.md" >}})