import { Position, Toaster } from "@blueprintjs/core";

/** Singleton toaster instance. Create separate instances for different options. */
const AppToaster = Toaster.create({
    position: Position.TOP
});

export const showErrorMessage = (error: Error|string) => {
    return AppToaster.show({icon: "error", intent: "danger", message: (error as Error).message || error});
};

export const showSuccessMessage = (message: string) => {
    return AppToaster.show({icon: "endorsed", intent: "success", message});
};
