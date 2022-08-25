import { Alert, Button } from "@mui/material";
import * as React from "react";
import { useDispatch, useSelector } from "react-redux";
import { dismissMessage } from "../../state/actions/messages-actions";
import { IconRepoState } from "../../state/reducers/root-reducer";

export const AppMessageList = () => {

	const messages = useSelector((state: IconRepoState) => state.messages);
	
	const dispatch = useDispatch();

	return messages?.idToMessageMap ? <div className="app-message-list">
		{
			Object.entries(messages?.idToMessageMap).map(([errorId, msg]) => (
				<div key={errorId}>
					<Alert
						severity={msg.error ? "error" : "info"}
						action={
							<Button color="inherit" size="small" onClick={() => dispatch(dismissMessage(errorId))}>
								Dismiss
							</Button>
						}
					>
						{(msg?.error || msg.info).toString()}
					</Alert>
				</div>
			))
		}
	</div> : null;
};
