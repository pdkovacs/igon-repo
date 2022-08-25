import "normalize.css";

import { useEffect } from "react";
import * as ReactDOM from "react-dom";
import { Provider as Redux, useDispatch } from "react-redux";
import { ErrorBoundary } from "react-error-boundary";
import store from "./state/store";
import { reportError } from "./state/actions/messages-actions";
import * as React from "react";
import { App } from "./app";

const ErrorFallback = ({ error }: { error: Error; }): JSX.Element|null => {
	const dispatch = useDispatch();

	useEffect(() => {
		dispatch(reportError(error));
	}, [error]);

  return null;
};

ReactDOM.render(
	<Redux store={store}>
		<ErrorBoundary
			FallbackComponent={ErrorFallback}
			onReset={() => undefined}
		>
			<App/>
		</ErrorBoundary>
	</Redux>,
	document.getElementById("app")
);
