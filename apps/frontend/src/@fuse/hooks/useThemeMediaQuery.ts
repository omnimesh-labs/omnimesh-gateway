'use client';
import { useEffect, useState } from 'react';
import { useTheme, Theme } from '@mui/material/styles';

/**
 * The useThemeMediaQuery function is a custom hook that returns a boolean indicating whether the current screen matches the specified media query.
 * It takes in a themeCallbackFunc as a parameter, which is a function that returns a string representing the media query to match.
 * It returns a boolean indicating whether the current screen matches the specified media query.
 */
function useThemeMediaQuery(themeCallbackFunc: (theme: Theme) => string) {
	const theme = useTheme();

	const query = themeCallbackFunc(theme).replace('@media ', '');

	/**
	 * The getMatches function checks if the current screen matches the specified media query.
	 * It takes in a media query string as a parameter and returns a boolean indicating whether the screen matches the query.
	 *
	 */
	function getMatches(q: string) {
		if (typeof window !== 'undefined') {
			return window.matchMedia(q).matches;
		}

		return false;
	}

	const [matches, setMatches] = useState(getMatches(query));

	useEffect(
		() => {
			const mediaQuery = window.matchMedia(query);

			// Update the state with the current value
			setMatches(getMatches(query));

			// Create an event listener
			const handler = (event: MediaQueryListEvent) => setMatches(event.matches);

			// Attach the event listener to know when the matches value changes
			mediaQuery.addEventListener('change', handler);

			// Remove the event listener on cleanup
			return () => mediaQuery.removeEventListener('change', handler);
		},
		[query] // Empty array ensures effect is only run on mount and unmount
	);

	return matches;
}

export default useThemeMediaQuery;
