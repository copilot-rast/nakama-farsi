import csv, sys, collections
f = sys.argv[1]; t0 = int(sys.argv[2]) if len(sys.argv) > 2 else 0
rows = [r for r in csv.DictReader(open(f))]
by = collections.defaultdict(list)
for r in rows: by[r['node']].append(r)
for n, rs in sorted(by.items()):
    hold = rs[-max(1, len(rs)//3):]
    mx = lambda k: max(float(r[k] or 0) for r in rs)
    av = lambda k, s: sum(float(r[k] or 0) for r in s)/len(s)
    print(f"{n}: rss_max={mx('rss_mb'):.0f}MB heap_max={mx('heap_mb'):.0f}MB cpu_hold_avg={av('cpu', hold):.0f}% cpu_max={mx('cpu'):.0f}% gor_max={mx('goroutines'):.0f} sess_max={mx('sessions'):.0f} pres_max={mx('presences'):.0f} redis_mb_max={mx('redis_mb')} redis_ops_hold={av('redis_ops', hold):.0f}")
