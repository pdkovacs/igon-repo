import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from "@mui/material";
import React from "react";
import { useDispatch } from "react-redux";
import { loginNeeded } from "../state/actions/app-actions";
import getEndPointUrl from "../services/url";

interface LoginDialogProps {
	readonly open:     boolean;
	readonly loginUrl: string;
}

export const LoginDialog = (props: LoginDialogProps) => {

	const dispatch = useDispatch();

	return <Dialog
		open={props.open}
	>
		<DialogTitle>
			Login
		</DialogTitle>
		<DialogContent>
			You have to be login again
		</DialogContent>
		<DialogActions>
			<Button
				// href={pathOfSelectedIconfile}
				onClick={() => {
					dispatch(loginNeeded(false));
					window.location.href = getEndPointUrl("/login");
				}}
			>
				OK
			</Button>
		</DialogActions>
	</Dialog>;
};
