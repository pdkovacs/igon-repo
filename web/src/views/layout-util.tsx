import * as React from "react";
import { Map } from "immutable";

import "./layout-util.styl";

export const renderMapAsTable = (properties: Map<string, JSX.Element>) =>
    <table className="property-list">
    <tbody>
        {properties.keySeq().map(k => <tr key={k}><td>{k}</td><td>{properties.get(k)}</td></tr>)}
    </tbody>
    </table>;
