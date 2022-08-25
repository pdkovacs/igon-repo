import { Alert, Button } from "@mui/material";
import * as React from "react";
import { useDispatch, useSelector } from "react-redux";
import { dismissNotification } from "../../state/actions/notification-actions";
import { IconRepoState } from "../../state/reducers/root-reducer";

export const NotificationList = () => {

	const notifications = useSelector((state: IconRepoState) => state.notifications);
	
	const dispatch = useDispatch();

	return notifications?.notificationId2NotificationMap ? <>
		{
			Object.entries(notifications?.notificationId2NotificationMap).map(([errorId, err]) => (
				<div key={errorId}>
					<Alert
						severity="error"
						action={
							<Button color="inherit" size="small" onClick={() => dispatch(dismissNotification(errorId))}>
								Dismiss
							</Button>
						}
					>
						{err.toString()}
					</Alert>
				</div>
			))
		}
	</> : null;
};
