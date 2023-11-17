import { receiveNotification } from "../state/actions/notification-actions";
import store from "../state/store";
import { getRealPath } from "./fetch-utils";

function dial() {

  const realPath = getRealPath("/subscribe");

	const conn = new WebSocket(`ws://${location.host}${realPath}`);

	conn.addEventListener("close", ev => {
		console.log(`WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`, true);
		if (ev.code !== 1001) {
			console.log("Reconnecting in 5s", true);
			setTimeout(dial, 5000); 
		}
	});
	
	conn.addEventListener("open", () => {
		console.info("websocket connected");
	});

	// This is where we handle messages received.
	conn.addEventListener("message", ev => {
		if (typeof ev.data !== "string") {
			console.error("unexpected message type", typeof ev.data);
			return;
		}
		console.log(ev.data);
		store.dispatch(receiveNotification(ev.data));
	});
}

console.log("calling dial()...");

dial();
