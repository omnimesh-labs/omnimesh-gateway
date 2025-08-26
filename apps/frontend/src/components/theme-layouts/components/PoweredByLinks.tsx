import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import { motion } from 'motion/react';
import Box from '@mui/material/Box';

/**
 * The powered by links.
 */
function PoweredByLinks() {
	const container = {
		show: {
			transition: {
				staggerChildren: 0.04
			}
		}
	};

	const item = {
		hidden: { opacity: 0, scale: 0.6 },
		show: { opacity: 1, scale: 1 }
	};

	return (
		<Box
			component={motion.div}
			variants={container}
			initial="hidden"
			animate="show"
			className="flex items-center gap-1 overflow-hidden"
		>
			<Tooltip
				title="React"
				placement="top"
			>
				<IconButton
					className="flex w-8 items-center justify-center rounded-none"
					component={motion.a}
					variants={item}
					href="https://reactjs.org/"
					target="_blank"
					rel="noreferrer noopener"
					role="button"
				>
					<img
						src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9Ii0xMS41IC0xMC4yMzE3NCAyMyAyMC40NjM0OCI+CiAgPHRpdGxlPlJlYWN0IExvZ288L3RpdGxlPgogIDxjaXJjbGUgY3g9IjAiIGN5PSIwIiByPSIyLjA1IiBmaWxsPSIjNjFkYWZiIi8+CiAgPGcgc3Ryb2tlPSIjNjFkYWZiIiBzdHJva2Utd2lkdGg9IjEiIGZpbGw9Im5vbmUiPgogICAgPGVsbGlwc2Ugcng9IjExIiByeT0iNC4yIi8+CiAgICA8ZWxsaXBzZSByeD0iMTEiIHJ5PSI0LjIiIHRyYW5zZm9ybT0icm90YXRlKDYwKSIvPgogICAgPGVsbGlwc2Ugcng9IjExIiByeT0iNC4yIiB0cmFuc2Zvcm09InJvdGF0ZSgxMjApIi8+CiAgPC9nPgo8L3N2Zz4K"
						alt="react"
						className="min-w-6"
					/>
				</IconButton>
			</Tooltip>
			<Tooltip
				title="Tanstack Query"
				placement="top"
			>
				<IconButton
					className="flex w-8 items-center justify-center rounded-none"
					component={motion.a}
					variants={item}
					href="https://tanstack.com/query/latest/docs/framework/react/overview"
					target="_blank"
					rel="noreferrer noopener"
					role="button"
				>
					<img
						src="data:image/png;base64,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"
						alt="material ui"
						className="min-w-6"
					/>
				</IconButton>
			</Tooltip>
			<Tooltip
				title="Tailwind"
				placement="top"
			>
				<IconButton
					className="flex w-8 items-center justify-center rounded-none"
					component={motion.a}
					variants={item}
					href="https://tailwindcss.com"
					target="_blank"
					rel="noreferrer noopener"
					role="button"
				>
					<span
						className="min-w-7"
						dangerouslySetInnerHTML={{
							__html: `
                <svg style='height: 0; width: 0; position: absolute; visibility: hidden;'>
                  <defs>
                    <linearGradient x1='0%' y1='0%' y2='100%' id='logoGradient'>
                      <stop stop-color='#2383AE' offset='0%'></stop>
                      <stop stop-color='#6DD7B9' offset='100%'></stop>
                    </linearGradient>
                  </defs>
                </svg>
                <svg class='w-full block' viewBox='0 0 64 64' xmlns='http://www.w3.org/2000/svg'><path d='M13.5 11.1C15.3 3.9 19.8.3 27 .3c10.8 0 12.15 8.1 17.55 9.45 3.6.9 6.75-.45 9.45-4.05-1.8 7.2-6.3 10.8-13.5 10.8-10.8 0-12.15-8.1-17.55-9.45-3.6-.9-6.75.45-9.45 4.05zM0 27.3c1.8-7.2 6.3-10.8 13.5-10.8 10.8 0 12.15 8.1 17.55 9.45 3.6.9 6.75-.45 9.45-4.05-1.8 7.2-6.3 10.8-13.5 10.8-10.8 0-12.15-8.1-17.55-9.45-3.6-.9-6.75.45-9.45 4.05z' transform='translate(5 16)' fill='url(#logoGradient)' fill-rule='evenodd'></path></svg>
                `
						}}
					/>
				</IconButton>
			</Tooltip>
			<Tooltip
				title="Next.js"
				placement="top"
			>
				<IconButton
					className="flex w-8 items-center justify-center rounded-none"
					component={motion.a}
					variants={item}
					href="https://nextjs.org"
					target="_blank"
					rel="noreferrer noopener"
					role="button"
				>
					<img
						src="/assets/images/logo/nextjs.svg"
						alt="next.js"
						className="min-w-6"
					/>
				</IconButton>
			</Tooltip>
			<Tooltip
				title="Auth.js"
				placement="top"
			>
				<IconButton
					className="flex w-8 items-center justify-center rounded-none"
					component={motion.a}
					variants={item}
					href="https://authjs.dev/"
					target="_blank"
					rel="noreferrer noopener"
					role="button"
				>
					<img
						src="/assets/images/logo/authjs.webp"
						alt="next.js"
						className="min-w-5"
					/>
				</IconButton>
			</Tooltip>
			<Tooltip
				title="TypeScript"
				placement="top"
			>
				<IconButton
					className="flex w-8 items-center justify-center"
					component={motion.a}
					variants={item}
					href="https://www.typescriptlang.org/"
					target="_blank"
					rel="noreferrer noopener"
					role="button"
				>
					<Box
						className="flex h-6 w-6 min-w-6 items-end justify-end overflow-hidden rounded-sm px-1 py-0.25"
						sx={{ backgroundColor: '#2e79c7!important', color: '#ffffff!important' }}
					>
						<span className="react-text text-sm text-xs font-semibold">TS</span>
					</Box>
				</IconButton>
			</Tooltip>
		</Box>
	);
}

export default PoweredByLinks;
