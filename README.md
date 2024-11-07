# goDumper
A simple script to dump process memory for the Linux os in addition to embedded systems like MikroTik. goDumber provides targeted dumps or full memory dumps.

# Usage 
- This simple script has two modes, targeted dump of a specific memory region like the process heap or stack, or full dump where goDumper will attempt to dump out the entire memory space of the process
## Full Dump
- Full process memory dump: use `-p <pid`
````
./godumper -p 1226246
[+] goDumper started
[+] Target PID: 1226246
[!] Skipping, failed to read memory: 7ffe98081000-7ffe98085000: read /proc/1226246/mem: input/output error
[+] Successful memory dump for pid: 1226246
````
- It is common to see certain areas of memory that are not possible to dump. goDumper will simply skip over these sections and alert you that they were skipped. goDumper will provided the memory range it was unable to read allowing the user to inspect skipped areas. See the FAQ section at the bottom for some reasons certain memory regions will be inaccessible.
- Verify the region that was skipped over to ensure it wasnt a vital area like `[stack]` or `[heap]`
````
cat /proc/1226246/maps
--snip--
7675b9b64000-7675b9b72000 r--p 0001f000 103:02 7892597                   /usr/lib/x86_64-linux-gnu/libtinfo.so.6.3
7675b9b72000-7675b9b76000 r--p 0002c000 103:02 7892597                   /usr/lib/x86_64-linux-gnu/libtinfo.so.6.3
7675b9b76000-7675b9b77000 rw-p 00030000 103:02 7892597                   /usr/lib/x86_64-linux-gnu/libtinfo.so.6.3
7675b9b90000-7675b9b97000 r--s 00000000 103:02 7866848                   /usr/lib/x86_64-linux-gnu/gconv/gconv-modules.cache
7675b9b97000-7675b9b99000 rw-p 00000000 00:00 0 
7675b9b99000-7675b9b9b000 r--p 00000000 103:02 7866467                   /usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2
7675b9b9b000-7675b9bc5000 r-xp 00002000 103:02 7866467                   /usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2
7675b9bc5000-7675b9bd0000 r--p 0002c000 103:02 7866467                   /usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2
7675b9bd1000-7675b9bd3000 r--p 00037000 103:02 7866467                   /usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2
7675b9bd3000-7675b9bd5000 rw-p 00039000 103:02 7866467                   /usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2
7ffe9803f000-7ffe98060000 rw-p 00000000 00:00 0                          [stack]
7ffe98081000-7ffe98085000 r--p 00000000 00:00 0                          [vvar]                                            <-- The memory region we failed to access, expected
7ffe98085000-7ffe98087000 r-xp 00000000 00:00 0                          [vdso]
ffffffffff600000-ffffffffff601000 --xp 00000000 00:00 0                  [vsyscall]
````
## Targeted Dump
- To conduct a targeted dump specify the target pid with `-p` and specify the range of memory with `-r` followed by the memory range.
- Get the memory range by looking at the processe's `maps` file under the `/proc` directory
````
cat /proc/1226246/maps
--snip--
7ffe9803f000-7ffe98060000 rw-p 00000000 00:00 0                          [stack]          <-- Our target range to dump
7ffe98081000-7ffe98085000 r--p 00000000 00:00 0                          [vvar]
7ffe98085000-7ffe98087000 r-xp 00000000 00:00 0                          [vdso]
ffffffffff600000-ffffffffff601000 --xp 00000000 00:00 0                  [vsyscall]
````
- Now we can provide the memory range to goDumper. This is the memory range of the processe's stack
````
./godumper -p 1226246 -r 7ffe9803f000-7ffe98060000
[+] GOLE started
[+] Target PID: 1226246
[+] Successful memory dump for pid: 1226246
````
- Dump files will be created in the `pwd` of the `goDumper` binary and will be named `dump.<pid>`. For example the resulting name for the above dump is `dump.1226246`

## FAQ
- **When conducting a memory dump with this simple script, why do certain memory sections fail?**
  - Kernel-Level Protections: Even if the memory regions have read permissions, certain kernel-level protections can prevent direct access. For instance, some memory regions (particularly those related to device files, like /dev/jiffies, or specific shared memory regions) might have restrictions that make them inaccessible to non-root users or from user space.
  - Dynamic Memory and Mapped File Changes: Processes may dynamically alter their memory mappings during execution. If memory mappings change between reading /proc/[pid]/maps and trying to access /proc/[pid]/mem, you might end up trying to read memory that’s no longer mapped or has changed permissions, resulting in an input/output error.
  - File-backed Memory Regions: Some entries in the maps file refer to memory-mapped files (like libc.so or libuc++.so). Even with read permissions, accessing these might occasionally result in errors due to how file-backed memory regions are managed by the kernel.
  - Physical or Virtual Memory Boundaries: When you attempt to read across memory boundaries, especially at the end of a segment, you can encounter input/output errors. This might be happening if, for example, there’s some overlap in addresses between memory segments, or if the range isn’t perfectly aligned.
