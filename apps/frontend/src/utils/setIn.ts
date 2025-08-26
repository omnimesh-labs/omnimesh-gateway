import { set, cloneDeep } from './lodashReplacements';

type State = Record<string, unknown> | Record<string, unknown>[];

const setIn = (state: State, name: string, value: unknown): State => {
	const clonedState = cloneDeep(state);
	return set(clonedState, name, value);
};

export default setIn;
