'use client';

import Typography from '@mui/material/Typography';
import clsx from 'clsx';
import { addDays, differenceInSeconds, isDate, parseISO } from 'date-fns';
import { memo, useCallback, useEffect, useRef, useState } from 'react';

type CountdownProps = {
	onComplete?: () => void;
	endDate?: Date | string;
	className?: string;
};

/**
 * Countdown
 * A React component used to display the number of days, hours, minutes and seconds left until a specified end date.
 * It allows a callback function to be passed in to be executed when the end date is reached.
 */
function Countdown(props: CountdownProps) {
	const { onComplete, endDate = addDays(new Date(), 15), className } = props;

	const [endDateVal] = useState(() => {
		if (isDate(endDate)) {
			return endDate;
		}

		if (typeof endDate === 'string') {
			return parseISO(endDate);
		}

		return addDays(new Date(), 15);
	});

	const [countdown, setCountdown] = useState({
		days: 0,
		hours: 0,
		minutes: 0,
		seconds: 0
	});
	const intervalRef = useRef<number | null>(null);

	const complete = useCallback(() => {
		if (intervalRef.current) {
			window.clearInterval(intervalRef.current);
		}

		if (onComplete) {
			onComplete();
		}
	}, [onComplete]);

	const tick = useCallback(() => {
		const currDate = new Date();
		const diff = differenceInSeconds(endDateVal, currDate);

		if (diff < 0) {
			complete();
			return;
		}

		const days = Math.floor(diff / (60 * 60 * 24));
		const hours = Math.floor((diff % (60 * 60 * 24)) / (60 * 60));
		const minutes = Math.floor((diff % (60 * 60)) / 60);
		const seconds = diff % 60;

		setCountdown({
			days,
			hours,
			minutes,
			seconds
		});
	}, [complete, endDateVal]);

	useEffect(() => {
		intervalRef.current = window.setInterval(tick, 1000);
		tick();
		return () => {
			if (intervalRef.current) {
				clearInterval(intervalRef.current);
			}
		};
	}, [tick]);

	return (
		<div className={clsx('flex items-center', className)}>
			<div className="flex flex-col items-center justify-center px-3">
				<Typography
					variant="h4"
					className="mb-1"
				>
					{countdown.days}
				</Typography>
				<Typography
					variant="caption"
					color="text.secondary"
				>
					days
				</Typography>
			</div>
			<div className="flex flex-col items-center justify-center px-3">
				<Typography
					variant="h4"
					className="mb-1"
				>
					{countdown.hours}
				</Typography>
				<Typography
					variant="caption"
					color="text.secondary"
				>
					hours
				</Typography>
			</div>
			<div className="flex flex-col items-center justify-center px-3">
				<Typography
					variant="h4"
					className="mb-1"
				>
					{countdown.minutes}
				</Typography>
				<Typography
					variant="caption"
					color="text.secondary"
				>
					minutes
				</Typography>
			</div>
			<div className="flex flex-col items-center justify-center px-3">
				<Typography
					variant="h4"
					className="mb-1"
				>
					{countdown.seconds}
				</Typography>
				<Typography
					variant="caption"
					color="text.secondary"
				>
					seconds
				</Typography>
			</div>
		</div>
	);
}

export default memo(Countdown);
