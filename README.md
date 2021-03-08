# OPA Persistent Experiment

## 10,000,000 items; Querying same item

This experiment loaded 10M items and then ran OPA queries for the same tenant repeatedly.

Using `inmem.store`

```
Alloc = 11565 MiB       TotalAlloc = 16651 MiB  Sys = 12352 MiB NumGC = 13
Alloc = 12137 MiB       TotalAlloc = 17223 MiB  Sys = 12960 MiB NumGC = 13
Alloc = 12701 MiB       TotalAlloc = 17787 MiB  Sys = 13568 MiB NumGC = 13
Alloc = 13280 MiB       TotalAlloc = 18366 MiB  Sys = 14176 MiB NumGC = 13
Alloc = 13857 MiB       TotalAlloc = 18943 MiB  Sys = 14785 MiB NumGC = 13
Alloc = 14426 MiB       TotalAlloc = 19512 MiB  Sys = 15392 MiB NumGC = 13
Alloc = 14890 MiB       TotalAlloc = 19976 MiB  Sys = 15866 MiB NumGC = 13
Alloc = 15049 MiB       TotalAlloc = 20135 MiB  Sys = 16066 MiB NumGC = 13
Alloc = 8181 MiB        TotalAlloc = 20548 MiB  Sys = 16139 MiB NumGC = 14
Alloc = 6376 MiB        TotalAlloc = 21072 MiB  Sys = 16146 MiB NumGC = 14
Alloc = 6816 MiB        TotalAlloc = 21512 MiB  Sys = 16149 MiB NumGC = 14
Alloc = 7275 MiB        TotalAlloc = 21971 MiB  Sys = 16151 MiB NumGC = 14
Alloc = 7721 MiB        TotalAlloc = 22417 MiB  Sys = 16153 MiB NumGC = 14
Alloc = 8180 MiB        TotalAlloc = 22876 MiB  Sys = 16155 MiB NumGC = 14
Alloc = 8629 MiB        TotalAlloc = 23325 MiB  Sys = 16158 MiB NumGC = 14
Alloc = 9079 MiB        TotalAlloc = 23775 MiB  Sys = 16160 MiB NumGC = 14
Alloc = 9524 MiB        TotalAlloc = 24220 MiB  Sys = 16162 MiB NumGC = 14
Alloc = 9975 MiB        TotalAlloc = 24671 MiB  Sys = 16164 MiB NumGC = 14
Alloc = 10426 MiB       TotalAlloc = 25122 MiB  Sys = 16167 MiB NumGC = 14
Alloc = 10768 MiB       TotalAlloc = 25464 MiB  Sys = 16168 MiB NumGC = 14
Alloc = 10941 MiB       TotalAlloc = 25637 MiB  Sys = 16168 MiB NumGC = 14
```

Using `persistent.store` w/ write load on startup:

```
Alloc = 122 MiB TotalAlloc = 22238 MiB  Sys = 2130 MiB  NumGC = 99
Alloc = 180 MiB TotalAlloc = 22796 MiB  Sys = 2130 MiB  NumGC = 104
Alloc = 150 MiB TotalAlloc = 23364 MiB  Sys = 2130 MiB  NumGC = 110
Alloc = 171 MiB TotalAlloc = 23921 MiB  Sys = 2130 MiB  NumGC = 116
Alloc = 181 MiB TotalAlloc = 24493 MiB  Sys = 2130 MiB  NumGC = 121
Alloc = 165 MiB TotalAlloc = 25078 MiB  Sys = 2130 MiB  NumGC = 127
Alloc = 145 MiB TotalAlloc = 25658 MiB  Sys = 2130 MiB  NumGC = 133
Alloc = 130 MiB TotalAlloc = 26244 MiB  Sys = 2130 MiB  NumGC = 139
Alloc = 112 MiB TotalAlloc = 26827 MiB  Sys = 2130 MiB  NumGC = 145
Alloc = 196 MiB TotalAlloc = 27411 MiB  Sys = 2130 MiB  NumGC = 150
Alloc = 180 MiB TotalAlloc = 27995 MiB  Sys = 2130 MiB  NumGC = 156
Alloc = 163 MiB TotalAlloc = 28578 MiB  Sys = 2130 MiB  NumGC = 162
Alloc = 148 MiB TotalAlloc = 29165 MiB  Sys = 2130 MiB  NumGC = 168
Alloc = 132 MiB TotalAlloc = 29749 MiB  Sys = 2130 MiB  NumGC = 174
Alloc = 118 MiB TotalAlloc = 30336 MiB  Sys = 2130 MiB  NumGC = 180
Alloc = 203 MiB TotalAlloc = 30921 MiB  Sys = 2130 MiB  NumGC = 185
Alloc = 183 MiB TotalAlloc = 31502 MiB  Sys = 2130 MiB  NumGC = 191
Alloc = 166 MiB TotalAlloc = 32086 MiB  Sys = 2130 MiB  NumGC = 197
Alloc = 142 MiB TotalAlloc = 32663 MiB  Sys = 2130 MiB  NumGC = 203
Alloc = 116 MiB TotalAlloc = 33237 MiB  Sys = 2130 MiB  NumGC = 209
Alloc = 198 MiB TotalAlloc = 33819 MiB  Sys = 2130 MiB  NumGC = 214
```

> The `./data` directory contained ~1.1G after compaction.

Using `persistent.store` w/o write load on startup:

```
Alloc = 117 MiB TotalAlloc = 1161 MiB   Sys = 741 MiB   NumGC = 7
Alloc = 170 MiB TotalAlloc = 1712 MiB   Sys = 741 MiB   NumGC = 12
Alloc = 142 MiB TotalAlloc = 2284 MiB   Sys = 742 MiB   NumGC = 18
Alloc = 125 MiB TotalAlloc = 2868 MiB   Sys = 742 MiB   NumGC = 24
Alloc = 198 MiB TotalAlloc = 3448 MiB   Sys = 742 MiB   NumGC = 30
Alloc = 193 MiB TotalAlloc = 4035 MiB   Sys = 742 MiB   NumGC = 35
Alloc = 175 MiB TotalAlloc = 4618 MiB   Sys = 742 MiB   NumGC = 41
Alloc = 157 MiB TotalAlloc = 5200 MiB   Sys = 742 MiB   NumGC = 47
Alloc = 140 MiB TotalAlloc = 5783 MiB   Sys = 742 MiB   NumGC = 53
Alloc = 122 MiB TotalAlloc = 6366 MiB   Sys = 742 MiB   NumGC = 59
Alloc = 205 MiB TotalAlloc = 6949 MiB   Sys = 742 MiB   NumGC = 64
Alloc = 186 MiB TotalAlloc = 7530 MiB   Sys = 742 MiB   NumGC = 70
Alloc = 170 MiB TotalAlloc = 8115 MiB   Sys = 742 MiB   NumGC = 76
Alloc = 147 MiB TotalAlloc = 8693 MiB   Sys = 742 MiB   NumGC = 82
Alloc = 129 MiB TotalAlloc = 9274 MiB   Sys = 742 MiB   NumGC = 88
Alloc = 115 MiB TotalAlloc = 9860 MiB   Sys = 742 MiB   NumGC = 94
Alloc = 199 MiB TotalAlloc = 10445 MiB  Sys = 742 MiB   NumGC = 99
Alloc = 179 MiB TotalAlloc = 11026 MiB  Sys = 742 MiB   NumGC = 105
Alloc = 156 MiB TotalAlloc = 11602 MiB  Sys = 742 MiB   NumGC = 111
Alloc = 139 MiB TotalAlloc = 12186 MiB  Sys = 742 MiB   NumGC = 117
Alloc = 121 MiB TotalAlloc = 12769 MiB  Sys = 742 MiB   NumGC = 123
```

> Tables took ~500ms to read on startup.


## 10,000,000 items; Querying 10,000 different items

Using `inmem.store`:

```
Alloc = 11543 MiB       TotalAlloc = 16629 MiB  Sys = 12286 MiB NumGC = 13
Alloc = 12099 MiB       TotalAlloc = 17185 MiB  Sys = 12893 MiB NumGC = 13
Alloc = 12653 MiB       TotalAlloc = 17739 MiB  Sys = 13501 MiB NumGC = 13
Alloc = 13208 MiB       TotalAlloc = 18294 MiB  Sys = 14109 MiB NumGC = 13
Alloc = 13765 MiB       TotalAlloc = 18851 MiB  Sys = 14650 MiB NumGC = 13
Alloc = 14319 MiB       TotalAlloc = 19405 MiB  Sys = 15258 MiB NumGC = 13
Alloc = 14769 MiB       TotalAlloc = 19855 MiB  Sys = 15731 MiB NumGC = 13
Alloc = 14973 MiB       TotalAlloc = 20059 MiB  Sys = 15932 MiB NumGC = 13
Alloc = 7754 MiB        TotalAlloc = 20468 MiB  Sys = 16072 MiB NumGC = 14
Alloc = 5879 MiB        TotalAlloc = 20971 MiB  Sys = 16079 MiB NumGC = 14
Alloc = 6295 MiB        TotalAlloc = 21387 MiB  Sys = 16081 MiB NumGC = 14
Alloc = 6735 MiB        TotalAlloc = 21827 MiB  Sys = 16083 MiB NumGC = 14
Alloc = 7169 MiB        TotalAlloc = 22261 MiB  Sys = 16086 MiB NumGC = 14
Alloc = 7613 MiB        TotalAlloc = 22705 MiB  Sys = 16088 MiB NumGC = 14
Alloc = 8049 MiB        TotalAlloc = 23140 MiB  Sys = 16090 MiB NumGC = 14
Alloc = 8476 MiB        TotalAlloc = 23568 MiB  Sys = 16092 MiB NumGC = 14
Alloc = 8917 MiB        TotalAlloc = 24009 MiB  Sys = 16094 MiB NumGC = 14
Alloc = 9360 MiB        TotalAlloc = 24452 MiB  Sys = 16096 MiB NumGC = 14
Alloc = 9794 MiB        TotalAlloc = 24886 MiB  Sys = 16099 MiB NumGC = 14
Alloc = 9993 MiB        TotalAlloc = 25084 MiB  Sys = 16099 MiB NumGC = 14
Alloc = 9004 MiB        TotalAlloc = 25356 MiB  Sys = 16099 MiB NumGC = 15
```

Using `persistent.store` w/o write load on startup:

```
Alloc = 170 MiB TotalAlloc = 1114 MiB   Sys = 675 MiB   NumGC = 6
Alloc = 164 MiB TotalAlloc = 1607 MiB   Sys = 676 MiB   NumGC = 11
Alloc = 171 MiB TotalAlloc = 2112 MiB   Sys = 676 MiB   NumGC = 16
Alloc = 195 MiB TotalAlloc = 2636 MiB   Sys = 676 MiB   NumGC = 21
Alloc = 123 MiB TotalAlloc = 3164 MiB   Sys = 676 MiB   NumGC = 27
Alloc = 150 MiB TotalAlloc = 3691 MiB   Sys = 676 MiB   NumGC = 32
Alloc = 183 MiB TotalAlloc = 4224 MiB   Sys = 676 MiB   NumGC = 37
Alloc = 125 MiB TotalAlloc = 4751 MiB   Sys = 676 MiB   NumGC = 43
Alloc = 146 MiB TotalAlloc = 5286 MiB   Sys = 676 MiB   NumGC = 48
Alloc = 178 MiB TotalAlloc = 5818 MiB   Sys = 676 MiB   NumGC = 53
Alloc = 202 MiB TotalAlloc = 6342 MiB   Sys = 676 MiB   NumGC = 58
Alloc = 128 MiB TotalAlloc = 6869 MiB   Sys = 676 MiB   NumGC = 64
Alloc = 156 MiB TotalAlloc = 7396 MiB   Sys = 676 MiB   NumGC = 69
Alloc = 184 MiB TotalAlloc = 7925 MiB   Sys = 676 MiB   NumGC = 74
Alloc = 109 MiB TotalAlloc = 8450 MiB   Sys = 676 MiB   NumGC = 80
Alloc = 129 MiB TotalAlloc = 8970 MiB   Sys = 676 MiB   NumGC = 85
Alloc = 161 MiB TotalAlloc = 9502 MiB   Sys = 676 MiB   NumGC = 90
Alloc = 166 MiB TotalAlloc = 10005 MiB  Sys = 676 MiB   NumGC = 95
Alloc = 190 MiB TotalAlloc = 10530 MiB  Sys = 676 MiB   NumGC = 100
Alloc = 183 MiB TotalAlloc = 11048 MiB  Sys = 676 MiB   NumGC = 106
Alloc = 139 MiB TotalAlloc = 11580 MiB  Sys = 676 MiB   NumGC = 111
```

* ~20x reduction in system memory usage
* ~60x reduction in heap usage
