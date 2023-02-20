const puppeteer = require("puppeteer");
const jwt = require('jsonwebtoken');

function signToken(){
    var token = jwt.sign({ email: 'admin@silly-snaps.com', role: 'admin', secret: process.env.FLAG, exp: Math.floor(Date.now() / 1000) + (60 * 60)}, process.env.JWT_KEY, { algorithm: 'HS256' });
    return token;
}

async function browse (url) {
    try {
        new URL(url);
    }
    catch(e){
        console.log("Invalid URL");
        return;
    }

    const browser = await puppeteer.launch({args: ["--no-sandbox"]});
    const page = await browser.newPage();

    var admin_token = signToken();
    const cookie = {name: "token", value: admin_token, domain: (process.env.DOMAIN || 'localhost')};

    await page.setCookie(cookie);
    
    await page.goto(url, {
        waitUntil: ["networkidle0", "domcontentloaded"],
      });

    res = await page.content();
    await browser.close();
}

browse(process.argv[2]);