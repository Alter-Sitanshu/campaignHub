import './footer.css';

const Footer = () => {
    const getYear = () => {
        let date = new Date;
        return date.getFullYear();
    };
    return (
        <footer className="footer">
            <div className="footer-container">
                <div className="footer-grid">
                {/* Logo + Copyright */}
                <div>
                    <a href="#" className="footer-logo">
                    FrogMedia
                    </a>
                    <p className="footer-copy">
                    © { getYear() } FrogMedia Inc. <br /> All rights reserved.
                    </p>
                </div>

                {/* Product Links */}
                <div>
                    <h4 className="footer-heading">Product</h4>
                    <ul className="footer-links">
                    <li><a href="#">For Creators</a></li>
                    <li><a href="#">For Brands</a></li>
                    <li><a href="#features">Features</a></li>
                    <li><a href="#pricing">Pricing</a></li>
                    </ul>
                </div>

                {/* Company Links */}
                <div>
                    <h4 className="footer-heading">Company</h4>
                    <ul className="footer-links">
                    <li><a href="#">About Us</a></li>
                    <li><a href="#">Careers</a></li>
                    <li><a href="#">Blog</a></li>
                    <li><a href="#">Contact</a></li>
                    </ul>
                </div>

                {/* Legal Links */}
                <div>
                    <h4 className="footer-heading">Legal</h4>
                    <ul className="footer-links">
                    <li><a href="#">Privacy Policy</a></li>
                    <li><a href="#">Terms of Service</a></li>
                    </ul>
                </div>
                </div>
            </div>
        </footer>
    );
}

export default Footer;