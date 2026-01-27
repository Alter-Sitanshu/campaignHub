import { useState, useEffect } from 'react';
import './landing_styles/hero.css';
import './landing_styles/features.css';
import './landing_styles/how_it_works.css';
import './landing_styles/comparison.css';
import './landing_styles/pricing.css';

import heroBg from '../../assets/hero-sample-2.jpg';

import Footer from '../../components/Footer/Footer';
import Navbar from '../../components/Navbar/Navbar';
import FeatureCard from '../../components/Card/FeatureCard';
import StepCard from '../../components/Card/StepCard';
import { useAuth } from '../../AuthContext';
import { useNavigate } from 'react-router-dom';
import { api } from "../../api.js";

const Landing = () => {
    const navigate = useNavigate();
    const [ coldStart, setColdStart ] = useState(true);

    let { user } = useAuth();
    useEffect(() => {
        let isMounted = true;

        const coldStart = async () => {
            try {
                await api.get("/"); // axios throws on non-2xx
                if (isMounted) {
                    setColdStart(false);
                }
            } catch (err) {
                if (isMounted) {
                    console.error("Cold start failed:", err);
                }
            }
        };

        coldStart();

        return () => {
            isMounted = false;
        };
    }, []);

    useEffect(() => {
        if(user && !coldStart) {
            navigate(`${user.entity}/dashboard/${user.id}`);
        }
    }, [user, coldStart])

    if (coldStart) {
        return (
            <div className="form-page">
                <div style={{ 
                    textAlign: 'center', paddingTop: '6rem', 
                    display: 'flex', flexDirection: 'column',
                    alignItems: 'center' 
                }}>
                    <p>Cold starting server. Please wait</p>
                    <div className="loading-dots">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <>
            <Navbar />
            <main>
                <div id="hero">
                    <div className='hero-text'>
                        <h1 id='main-text'>
                            Find Your Perfect Brand.
                            <span> Launch Your Next Campaign.</span>
                        </h1>
                        <p>
                            FrogMedia is the ultimate platform for creators and brands to build authentic, powerful, and profitable partnerships.
                        </p>
                        <div className='hero-button-group'>
                            <a href="/auth/accounts?entity=users" className='get-started-button'>
                                Get Started Free
                            </a>
                            <a href="/auth/accounts?entity=brands" className='for-brands-button'>
                                For Brands
                            </a>
                        </div>
                    </div>
                    <div className="hero-img">
                        <div className="hero-img-inner">
                            <img src={heroBg} alt="hero-bg" loading='lazy'/>
                        </div>
                    </div>
                </div>
                <section id="features" className="features-section">
                    <div className="features-container">
                        <div className="features-header">
                            <h2 className="features-title">The Platform for Authentic Partnerships</h2>
                            <p className="features-subtitle">
                                Stop searching. Start connecting. We provide the tools you need to build meaningful campaigns from start to finish.
                            </p>
                        </div>

                        <div className="features-grid">
                            {/* Feature 1 */}
                            <FeatureCard
                                title="ViewPay"
                                short="Get paid for every authenticate view you get. Transparent and seamless transactions."
                            />

                            {/* Feature 2 */}
                            <FeatureCard 
                                title="Seamless Collaboration"
                                short="Manage contracts, communicate, and approve content all in one place. No more endless email chains."
                            />

                            {/* Feature 3 */}
                            <FeatureCard
                                title="Transparent Analytics"
                                short="Track your campaign's performance in real-time with a clear, actionable data dashboard."
                            />
                        </div>
                    </div>
                </section>
                <section id="how-it-works" className="how-it-works">
                    <div className="hiw-container">
                        <div className="steps-headline-div">
                            <h2>How It Works</h2>
                            <p>
                                From discovery to delivery, we've streamlined every step of the partnership process.
                            </p>
                        </div>

                        <div className="steps-div">
                            <StepCard 
                                step={1}
                                head="Create Your Profile"
                                desc="Sign up and showcase your brand or creator portfolio. Add your niche, audience stats, and campaign preferences."
                            />
                            <StepCard 
                                step={2}
                                head="Discover and Connect"
                                desc="Browse through verified partners, send collaboration requests, and negotiate terms directly on the platform."
                            />
                            <StepCard 
                                step={3}
                                head="Launch & Track"
                                desc="Execute your campaign with built-in tools for content approval, payments, and real-time performance analytics."
                            />
                        </div>
                    </div>
                </section>
                <div className="separation">
                    <div className='bar'></div>
                </div>
                <section id="comparison" className="comparison">
                    <div className="comparison-container">
                        <h2 className="comparison-title">
                            Why Choose FrogMedia?
                        </h2>
                        <div className="comparison-grid">
                            <div className='previous'>
                                <h3>Traditional Approach</h3>
                                <ul className='comparison-list prev-list'>
                                    <li className='list-item'>
                                        <span className="icon-cross">✗</span>
                                        Endless email threads
                                    </li>
                                    <li className='list-item'>
                                        <span className="icon-cross">✗</span>
                                        No transparency
                                    </li>
                                    <li className='list-item'>
                                        <span className="icon-cross">✗</span>
                                        Manual tracking
                                    </li>
                                    <li className='list-item'>
                                        <span className="icon-cross">✗</span>
                                        Payment delays
                                    </li>
                                </ul>
                            </div>
                            <div className='after'>
                                <h3>With FrogMedia</h3>
                                <ul className='comparison-list after-list'>
                                    <li className='list-item'>
                                        <span className="icon-check">✓</span>
                                        Centralized communication
                                    </li>
                                    <li className='list-item'>
                                        <span className="icon-check">✓</span>
                                        Real-time analytics
                                    </li>
                                    <li className='list-item'>
                                        <span className="icon-check">✓</span>
                                        Automated workflows
                                    </li>
                                    <li className='list-item'>
                                        <span className="icon-check">✓</span>
                                        Secure instant payments
                                    </li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </section>
                <section id="pricing" className="pricing-section">
                    <div className="pricing-container">
                        <h2 className="pricing-title">
                        Ready to Grow Your Influence?
                        </h2>
                        <p className="pricing-subtitle">
                        Join thousands of top creators and innovative brands building their legacy on FrogMedia.
                        </p>
                        <a
                        href="/auth/accounts?entity=users"
                        className="pricing-button"
                        >
                        Sign Up Now — It's Free
                        </a>
                    </div>
                </section>
            </main>
            <Footer />
        </>
    );
};

export default Landing;