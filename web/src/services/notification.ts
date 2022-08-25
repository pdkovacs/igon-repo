import { receiveNotification } from "../state/actions/notification-actions";
import store from "./../state/store";

function dial() {
	const conn = new WebSocket(`ws://${location.host}/subscribe`);

	conn.addEventListener("close", ev => {
		console.log(`WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`, true);
		if (ev.code !== 1001) {
			console.log("Reconnecting in 1s", true);
			setTimeout(dial, 1000);
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

dial();
