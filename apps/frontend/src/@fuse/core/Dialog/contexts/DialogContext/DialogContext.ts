'use client';

import React from 'react';
import { DialogProps } from '../../Dialog';

export interface DialogContextType {
	dialogs: Record<string, DialogProps>;
	openDialog: (T: DialogProps) => void;
	closeDialog: (id: DialogProps['id']) => void;
}

export const DialogDefaultContext: DialogContextType = {
	dialogs: {},
	openDialog: () => null,
	closeDialog: () => null
};

// Dialog context
export const DialogContext = React.createContext<DialogContextType>(DialogDefaultContext);
