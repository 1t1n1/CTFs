import angr

proj = angr.Project("service")
simgr = proj.factory.simulation_manager()
simgr.explore(find=lambda s: b"successfully" in s.posix.dumps(1))

s = simgr.found[0]
key = s.posix.dumps(0)
print(f"Found pin {key.decode()}")