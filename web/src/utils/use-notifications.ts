import { useEffect, useRef } from "react";
import { useDispatch, useSelector } from "react-redux";
import { fetchIconsAction } from "../state/actions/icons-actions";
import { IconRepoState } from "../state/reducers/root-reducer";
import { useReporters } from "./use-reporters";

const changeInIconList = [
	"iconCreated",
	"iconDeleted",
	"iconfileAdded",
	"iconfileDeleted"
];

export const useNotifications = () => {

	const dispatch = useDispatch();

	const getNotificationParams = (notification: string): [string, string, () => void] => {
		if (changeInIconList.includes(notification)) {
			return [notification, "Refresh icon list", () => dispatch(fetchIconsAction())];
		} else {
			console.warn("Unexpected notification from backend: ", notification);
			return [notification, "Dismiss", (): void => undefined];
		}
	};
	
		const notifications = useSelector((state: IconRepoState) => state.notifications);
	const { reportNotification } = useReporters();

	const lastSeenNotification = useRef<number>(-1);

	useEffect(() => {
		const notificationMap = notifications.notificationId2NotificationMap; 
		Object.keys(notificationMap).forEach(
			noticationId => {
				const id = parseInt(noticationId, 10);
				if (isNaN(id)) {
					console.error("Notification id is not a number:", noticationId);
					return;
				} else 
				if (id > lastSeenNotification.current) {
					reportNotification(...getNotificationParams(notificationMap[id]));
				}
			}
		);
		lastSeenNotification.current = notifications.lastId;
	}, [notifications]);

};
