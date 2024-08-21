import classNames from 'classnames'
import styles from '../../Home/Home.module.scss'
import { PrimaryButton } from '../Buttons/PrimaryButton'
import { Typography } from '../Typography/Typography'

export const OSSCallToAction = () => {
	return (
		<div className={'flex justify-center mx-5 md:mx-[10vw] my-10 md:my-32'}>
			<div
				className={classNames(
					styles.ossCard,
					'w-full max-w-[1250px] border-[1px] border-divider-on-dark rounded-3xl py-10 px-2',
				)}
				style={{
					backgroundColor: '#30294E',
				}}
			>
				<h3 className="text-center leading-normal">
					Join our{' '}
					<span className={styles.highlightedText}>
						Fullstack Monitoring for .NET Applications
					</span>{' '}
					livestream.
				</h3>
				<div className="text-center px-2 md:px-16 mt-6">
					<Typography type="copy1" className="leading-relaxed">
						Join us September 5th at 1pm PDT to learn how to
						leverage OpenTelemetry to get fullstack visibility into
						everything happening in your .NET applications.
					</Typography>
				</div>
				<div className="flex justify-center mt-8">
					<div className="flex flex-col lg:flex-row justify-center gap-4 w-full px-5 md:w-auto">
						<PrimaryButton
							href="https://lu.ma/51zg6p43?utm_source=highlight-oss-cta"
							target="_blank"
							rel="noreferrer"
							className="md:max-w-[180px]"
						>
							<div className="flex justify-center items-center gap-3">
								<Typography type="copy2" emphasis={true}>
									Register Here
								</Typography>
							</div>
						</PrimaryButton>
					</div>
				</div>
			</div>
		</div>
	)
}
