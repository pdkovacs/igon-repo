import "normalize.css";

import { useEffect } from "react";
import * as ReactDOM from "react-dom";
import { Provider as Redux } from "react-redux";
import { ErrorBoundary } from "react-error-boundary";
import store from "./state/store";
import * as React from "react";
import { App } from "./app";
import { SnackbarProvider } from "notistack";
import { useReporters } from "./utils/use-reporters";

const ErrorFallback = ({ error }: { error: Error; }): JSX.Element|null => {

	const { reportError } = useReporters();

	useEffect(() => {
		reportError(error.message);
	}, [error]);

  return null;
};

ReactDOM.render(
	<Redux store={store}>
		<SnackbarProvider
			maxSnack={3}
		>
			<ErrorBoundary
				FallbackComponent={ErrorFallback}
				onReset={() => undefined}
			>
				<App/>
			</ErrorBoundary>
		</SnackbarProvider>
	</Redux>,
	document.getElementById("app")
);
