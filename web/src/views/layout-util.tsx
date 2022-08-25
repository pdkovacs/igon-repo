import * as React from "react";

import "./layout-util.styl";

export const renderMapAsTable = (properties: { [name: string]: JSX.Element }) =>
	<table className="property-list">
		<tbody>
			{Object.keys(properties).map(k => <tr key={k}><td className="property-name">{k}</td><td className="property-value">{properties[k]}</td></tr>)}
		</tbody>
	</table>;
