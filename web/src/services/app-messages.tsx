import React from "react";
import { Button } from "@mui/material";
import { SnackbarKey, useSnackbar } from "notistack";

export const useReporter = () => {
	const { enqueueSnackbar, closeSnackbar } = useSnackbar();

	const action = (snackbarId: SnackbarKey) => (
		<Button color="inherit" onClick={() => closeSnackbar(snackbarId)}>
			Dismiss
		</Button>
	);

	const reportError = (msg: string) => {
		enqueueSnackbar(msg, {
			variant: "error",
			persist: true,
			action
		});
	};

	const reportInfo = (msg: string) => {
		enqueueSnackbar(msg, {
			variant: "info",
			action
		});
	};	
	
	return {
		reportError,
		reportInfo
	};

};
