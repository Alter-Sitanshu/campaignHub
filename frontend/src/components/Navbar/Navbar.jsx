import './Navbar.css';

const Navbar = () => {
    const toggleMobileMenu = () => {
        const menu = document.getElementById('mobile-menu');
        menu.classList.toggle('active');
    };

    return (
        <header className="navbar-header">
            <nav className="navbar-container">
                <a href="#" className="navbar-logo">
                    FrogMedia
                </a>
                <div className="navbar-links">
                    <a href="#features" className="navbar-link">Features</a>
                    <a href="#how-it-works" className="navbar-link">How It Works</a>
                    <a href="#pricing" className="navbar-link">Pricing</a>
                    <a href="/auth/accounts?entity=users" className="navbar-cta">
                        Get Started
                    </a>
                </div>

                <button 
                    id="menu-btn" 
                    className="menu-button"
                    onClick={toggleMobileMenu}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="menu-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16m-7 6h7" />
                    </svg>
                </button>
            </nav>

            <div id="mobile-menu" className="mobile-menu">
                <a href="#features" className="mobile-menu-link">Features</a>
                <a href="#how-it-works" className="mobile-menu-link">How It Works</a>
                <a href="#pricing" className="mobile-menu-link">Pricing</a>
                <a href="/auth/accounts?entity=users" className="mobile-menu-cta">
                    Get Started
                </a>
            </div>
        </header>
    );
};

export default Navbar;