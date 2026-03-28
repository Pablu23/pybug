import bdb
import json
import sys
from types import FrameType


class PyBugBridgeDebugger(bdb.Bdb):
    def __init__(self):
        super().__init__()
        self.waiting = False

    def send(self, msg: dict):
        sys.stdout.write(json.dumps(msg) + "\n")
        sys.stdout.flush()

    def recv(self) -> dict:
        return json.loads(sys.stdin.readline())

    def user_line(self, frame: FrameType):
        print("TRACE:", frame.f_code.co_filename, frame.f_lineno)
        self.send(
            {
                "event": "stopped",
                "file": frame.f_code.co_filename,
                "line": frame.f_lineno,
            }
        )

        self.interaction_loop(frame)

    def interaction_loop(self, frame: FrameType):
        while True:
            print("DEBUG: waiting for command", file=sys.stderr, flush=True)
            cmd = self.recv()
            print(f"DEBUG: received command {cmd}", file=sys.stderr, flush=True)

            match cmd["cmd"]:
                case "continue":
                    return self.set_continue()
                case "step":
                    return self.set_step()
                case "next":
                    return self.set_next(frame)
                case "locals":
                    self.send(
                        {
                            "request_id": cmd["request_id"],
                            "event": "locals",
                            "vars": {k: repr(v) for k, v in frame.f_locals.items()},
                        }
                    )
                case "eval":
                    try:
                        result = eval(cmd["expr"], frame.f_globals, frame.f_locals)
                        self.send(
                            {
                                "request_id": cmd["request_id"],
                                "event": "eval",
                                "status": "ok",
                                "value": repr(result),
                            }
                        )
                    except Exception as e:
                        self.send(
                            {
                                "request_id": cmd["request_id"],
                                "event": "eval",
                                "status": "error",
                                "value": str(e),
                            }
                        )
                case "break":
                    err = self.set_break(cmd["file"], cmd["line"])
                    if err:
                        self.send(
                            {
                                "request_id": cmd["request_id"],
                                "event": "break",
                                "status": "error",
                                "error": err,
                            }
                        )
                    else:
                        self.send(
                            {
                                "request_id": cmd["request_id"],
                                "event": "break",
                                "status": "ok",
                            }
                        )


def main():
    if len(sys.argv) < 2:
        print("Usage: pydbug_runtime <script.py>")

    script = sys.argv[1]
    sys.argv = sys.argv[1:]

    dbg = PyBugBridgeDebugger()

    with open(script, "rb") as f:
        code = compile(f.read(), script, "exec")

    globals_dict = {
        "__name__": "__main__",
        "__file__": script,
        "__package__": None,
    }

    dbg.set_step()
    dbg.run(code, globals_dict, globals_dict)


if __name__ == "__main__":
    main()
