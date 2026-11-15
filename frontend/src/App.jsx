import React, { useState, useEffect } from 'react';
import SaltwakeHero from './SaltwakeHero';
import { api } from './api';

export default function App() {
  const [activeModal, setActiveModal] = useState(null); // 'shop', 'auth', 'cart', 'admin', 'puzzle', 'my-orders'
  const [user, setUser] = useState(api.getUser());
  const [authMode, setAuthMode] = useState('login'); // 'login' or 'register'
  
  // Auth Form
  const [authEmail, setAuthEmail] = useState('');
  const [authPassword, setAuthPassword] = useState('');
  const [authName, setAuthName] = useState('');
  const [authRole, setAuthRole] = useState('user');
  const [authError, setAuthError] = useState('');

  // Products & Cart
  const [products, setProducts] = useState([]);
  const [cart, setCart] = useState([]); // [{ product, quantity }]

  // Puzzle Payment & User Orders
  const [puzzleData, setPuzzleData] = useState(null);
  const [puzzleAnswer, setPuzzleAnswer] = useState('');
  const [orderSuccess, setOrderSuccess] = useState(null);
  const [userOrders, setUserOrders] = useState([]);

  // Store Owner Dashboard
  const [adminProducts, setAdminProducts] = useState([]);
  const [adminOrders, setAdminOrders] = useState([]);
  const [newProdName, setNewProdName] = useState('');
  const [newProdPrice, setNewProdPrice] = useState('');
  const [newProdStock, setNewProdStock] = useState('');
  const [newProdDesc, setNewProdDesc] = useState('');
  const [editingProduct, setEditingProduct] = useState(null); // Product being edited

  useEffect(() => {
    loadProducts();
  }, []);

  const loadProducts = async () => {
    try {
      const data = await api.getProducts();
      setProducts(data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleAuth = async (e) => {
    e.preventDefault();
    setAuthError('');
    try {
      if (authMode === 'register') {
        await api.register({
          name: authName,
          email: authEmail,
          password: authPassword,
          role: authRole
        });
        alert('Registration successful! Please log in.');
        setAuthMode('login');
      } else {
        const res = await api.login({
          email: authEmail,
          password: authPassword
        });
        api.setAuth(res.token, res.user);
        setUser(res.user);
        setActiveModal(null);
      }
    } catch (err) {
      setAuthError(err.message || 'Authentication error');
    }
  };

  const handleLogout = () => {
    api.clearAuth();
    setUser(null);
    setActiveModal(null);
  };

  const addToCart = (prod) => {
    setCart((prev) => {
      const existing = prev.find((item) => item.product._id === prod._id);
      if (existing) {
        return prev.map((item) =>
          item.product._id === prod._id
            ? { ...item, quantity: item.quantity + 1 }
            : item
        );
      }
      return [...prev, { product: prod, quantity: 1 }];
    });
    alert(`Added ${prod.name} to cargo hold!`);
  };

  const handleStartCheckout = async () => {
    if (!user) {
      setActiveModal('auth');
      return;
    }
    if (cart.length === 0) return;

    try {
      const items = cart.map((c) => ({
        productId: c.product._id,
        quantity: c.quantity
      }));
      const res = await api.requestChallenge(items);
      setPuzzleData(res);
      setPuzzleAnswer('');
      setActiveModal('puzzle');
    } catch (err) {
      alert(err.message);
    }
  };

  const handleConfirmPuzzle = async (e) => {
    e.preventDefault();
    try {
      const items = cart.map((c) => ({
        productId: c.product._id,
        quantity: c.quantity
      }));
      const res = await api.confirmOrder(puzzleData.challengeId, puzzleAnswer, items);
      setOrderSuccess(res.order);
      setCart([]);
      loadProducts();
    } catch (err) {
      alert(err.message || 'Incorrect puzzle answer!');
    }
  };

  const openMyOrders = async () => {
    if (!user) {
      setActiveModal('auth');
      return;
    }
    try {
      const ords = await api.getMyOrders();
      setUserOrders(ords || []);
      setActiveModal('my-orders');
    } catch (err) {
      alert(err.message);
    }
  };

  const openAdmin = async () => {
    if (!user || user.role !== 'admin') {
      alert('Only Store Owners / Admins can access the Captain\'s Deck!');
      return;
    }
    try {
      const prods = await api.getAdminProducts();
      const ords = await api.getAdminOrders();
      setAdminProducts(prods || []);
      setAdminOrders(ords || []);
      setActiveModal('admin');
    } catch (err) {
      alert(err.message);
    }
  };

  const handleCreateProduct = async (e) => {
    e.preventDefault();
    try {
      if (editingProduct) {
        await api.updateProduct(editingProduct._id, {
          name: newProdName,
          price: parseFloat(newProdPrice),
          stock: parseInt(newProdStock, 10),
          description: newProdDesc
        });
        alert('Cargo item updated successfully!');
        setEditingProduct(null);
      } else {
        await api.createProduct({
          name: newProdName,
          price: parseFloat(newProdPrice),
          stock: parseInt(newProdStock, 10),
          description: newProdDesc
        });
        alert('Cargo item added to your inventory!');
      }
      setNewProdName('');
      setNewProdPrice('');
      setNewProdStock('');
      setNewProdDesc('');
      const prods = await api.getAdminProducts();
      setAdminProducts(prods || []);
      loadProducts();
    } catch (err) {
      alert(err.message);
    }
  };

  const startEditProduct = (prod) => {
    setEditingProduct(prod);
    setNewProdName(prod.name);
    setNewProdPrice(prod.price.toString());
    setNewProdStock(prod.stock.toString());
    setNewProdDesc(prod.description || '');
  };

  const cancelEditProduct = () => {
    setEditingProduct(null);
    setNewProdName('');
    setNewProdPrice('');
    setNewProdStock('');
    setNewProdDesc('');
  };

  const handleDeleteProduct = async (id) => {
    if (!confirm('Are you sure you want to remove this item?')) return;
    try {
      await api.deleteProduct(id);
      const prods = await api.getAdminProducts();
      setAdminProducts(prods || []);
      loadProducts();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleUpdateOrderStatus = async (orderId, newStatus) => {
    try {
      await api.updateOrderStatus(orderId, newStatus);
      const ords = await api.getAdminOrders();
      setAdminOrders(ords || []);
    } catch (err) {
      alert(err.message);
    }
  };

  // Stats calculation for store owner
  const totalItemsSold = adminOrders.reduce((sum, ord) => {
    return sum + ord.items.reduce((itemSum, it) => itemSum + it.quantity, 0);
  }, 0);
  const totalRevenue = adminOrders.reduce((sum, ord) => {
    return sum + ord.items.reduce((itemSum, it) => itemSum + (it.price * it.quantity), 0);
  }, 0);

  return (
    <>
      <SaltwakeHero
        onExploreShop={() => setActiveModal('shop')}
        onOpenCrew={() => {
          if (!user) {
            setActiveModal('auth');
          } else if (user.role === 'admin') {
            openAdmin();
          } else {
            openMyOrders();
          }
        }}
      />

      {/* Floating Action Bar */}
      <div style={{ position: 'fixed', bottom: 30, right: 30, zIndex: 60, display: 'flex', gap: 14, flexWrap: 'wrap' }}>
        <button className="pill" style={{ padding: '12px 24px', fontSize: 14 }} onClick={() => setActiveModal('shop')}>
          Bazaars ({products.length})
        </button>
        <button className="pill" style={{ padding: '12px 24px', fontSize: 14 }} onClick={() => setActiveModal('cart')}>
          Cargo Hold ({cart.reduce((s, i) => s + i.quantity, 0)})
        </button>
        {user ? (
          <>
            <button className="pill" style={{ padding: '12px 24px', fontSize: 14 }} onClick={openMyOrders}>
              My Purchases
            </button>
            {user.role === 'admin' && (
              <button className="pill" style={{ padding: '12px 24px', fontSize: 14, background: '#e3d6bc' }} onClick={openAdmin}>
                Captain's Deck
              </button>
            )}
            <button className="pill" style={{ padding: '12px 18px', fontSize: 13, background: '#f8d7da' }} onClick={handleLogout}>
              Logout
            </button>
          </>
        ) : (
          <button className="pill" style={{ padding: '12px 24px', fontSize: 14 }} onClick={() => setActiveModal('auth')}>
            Board / Login
          </button>
        )}
      </div>

      {/* SHOP / BAZAARS MODAL */}
      {activeModal === 'shop' && (
        <div className="modal-overlay" onClick={() => setActiveModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 840 }}>
            <div className="modal-header">
              <h2>Port Royal Bazaars</h2>
              <button className="modal-close" onClick={() => setActiveModal(null)}>✕</button>
            </div>
            <p style={{ letterSpacing: '0.05em', textTransform: 'uppercase', fontSize: 13, color: '#555' }}>
              Genuine trade goods inspected by licensed store captains
            </p>
            <div className="product-grid">
              {products.length === 0 ? (
                <p style={{ padding: 20 }}>No trade goods currently in port.</p>
              ) : (
                products.map((p) => (
                  <div key={p._id} className="product-card">
                    <div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span className="badge" style={{ background: p.stock > 0 ? '#d4edda' : '#f8d7da' }}>
                          Stock: {p.stock}
                        </span>
                        <strong style={{ fontSize: 18 }}>${p.price.toFixed(2)}</strong>
                      </div>
                      <h3 style={{ margin: '14px 0 6px', fontSize: 20, fontFamily: 'Goudy Bookletter 1911' }}>{p.name}</h3>
                      <p style={{ fontSize: 13, color: '#444', marginBottom: 16 }}>{p.description || 'Parchment trade item'}</p>
                    </div>
                    <button
                      className="pill"
                      style={{ padding: '10px 16px', fontSize: 13, width: '100%', textAlign: 'center' }}
                      disabled={p.stock <= 0}
                      onClick={() => addToCart(p)}
                    >
                      {p.stock > 0 ? 'Stow in Cargo' : 'Sold Out'}
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}

      {/* CARGO / CART MODAL */}
      {activeModal === 'cart' && (
        <div className="modal-overlay" onClick={() => setActiveModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Your Cargo Hold</h2>
              <button className="modal-close" onClick={() => setActiveModal(null)}>✕</button>
            </div>
            {cart.length === 0 ? (
              <p>Your cargo hold is empty. Visit the Bazaars to trade.</p>
            ) : (
              <div>
                {cart.map((item) => (
                  <div key={item.product._id} style={{ display: 'flex', justifyContent: 'space-between', padding: '12px 0', borderBottom: '1px solid #ccc' }}>
                    <div>
                      <strong>{item.product.name}</strong>
                      <div style={{ fontSize: 12, color: '#666' }}>Quantity: {item.quantity} × ${item.product.price}</div>
                    </div>
                    <strong>${(item.quantity * item.product.price).toFixed(2)}</strong>
                  </div>
                ))}
                <div style={{ marginTop: 20, display: 'flex', justifyContent: 'space-between', fontSize: 20, fontWeight: 'bold' }}>
                  <span>Total Cargo:</span>
                  <span>${cart.reduce((s, i) => s + i.quantity * i.product.price, 0).toFixed(2)}</span>
                </div>
                <button
                  className="pill"
                  style={{ width: '100%', marginTop: 24, padding: '14px 20px', textAlign: 'center' }}
                  onClick={handleStartCheckout}
                >
                  Pay & Unlock Cargo with Riddle
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* USER'S PURCHASES / MY ORDERS MODAL */}
      {activeModal === 'my-orders' && (
        <div className="modal-overlay" onClick={() => setActiveModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 750 }}>
            <div className="modal-header">
              <h2>My Purchases & Cargo Receipts</h2>
              <button className="modal-close" onClick={() => setActiveModal(null)}>✕</button>
            </div>
            {userOrders.length === 0 ? (
              <p>No past voyages or orders found.</p>
            ) : (
              <div>
                {userOrders.map((ord) => (
                  <div key={ord._id} style={{ border: '2px solid var(--line)', padding: 16, borderRadius: 12, marginBottom: 14, background: '#ece7d8' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <div>
                        <strong>Receipt #{ord._id.slice(0, 8)}</strong>
                        <div style={{ fontSize: 12, color: '#666' }}>Purchased on: {new Date(ord.createdAt).toLocaleString()}</div>
                      </div>
                      <span className="badge" style={{ background: '#d4edda' }}>Status: {ord.status}</span>
                    </div>
                    <div style={{ margin: '12px 0', paddingLeft: 8, borderLeft: '3px solid var(--line)' }}>
                      {ord.items.map((it, idx) => (
                        <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', fontSize: 14, margin: '4px 0' }}>
                          <span>• {it.productName} (Qty: {it.quantity})</span>
                          <span>${(it.price * it.quantity).toFixed(2)}</span>
                        </div>
                      ))}
                    </div>
                    <div style={{ textAlign: 'right', fontWeight: 'bold', fontSize: 16 }}>
                      Total Paid: ${ord.totalAmount.toFixed(2)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* PUZZLE PAYMENT MODAL */}
      {activeModal === 'puzzle' && puzzleData && (
        <div className="modal-overlay">
          <div className="modal-content" style={{ textAlign: 'center' }}>
            <div className="modal-header">
              <h2>Shipmaster's Riddle</h2>
              <button className="modal-close" onClick={() => setActiveModal(null)}>✕</button>
            </div>
            {!orderSuccess ? (
              <form onSubmit={handleConfirmPuzzle}>
                <p style={{ fontSize: 16, marginBottom: 20 }}>Answer the puzzle to release the shipment:</p>
                <div style={{ background: '#dcd7c7', padding: '24px', borderRadius: 16, border: '3px solid var(--line)', marginBottom: 20 }}>
                  <h3 style={{ fontSize: 26, fontFamily: 'Goudy Bookletter 1911' }}>{puzzleData.question}</h3>
                </div>
                <div className="form-group">
                  <input
                    type="text"
                    required
                    placeholder="Enter answer here..."
                    className="form-input"
                    value={puzzleAnswer}
                    onChange={(e) => setPuzzleAnswer(e.target.value)}
                    style={{ textAlign: 'center', fontSize: 18 }}
                  />
                </div>
                <button type="submit" className="pill" style={{ padding: '14px 40px', marginTop: 10 }}>
                  Confirm Order (${puzzleData.totalAmount.toFixed(2)})
                </button>
              </form>
            ) : (
              <div>
                <h3 style={{ color: '#2d6a4f', fontSize: 24, margin: '16px 0' }}>Order Sealed & Paid!</h3>
                <p>Order ID: {orderSuccess._id}</p>
                <p style={{ marginTop: 8 }}>Total Paid: ${orderSuccess.totalAmount.toFixed(2)}</p>
                <div style={{ display: 'flex', gap: 12, justifyContent: 'center', marginTop: 24 }}>
                  <button
                    className="pill"
                    style={{ padding: '10px 24px' }}
                    onClick={() => {
                      setOrderSuccess(null);
                      openMyOrders();
                    }}
                  >
                    View in My Purchases
                  </button>
                  <button
                    className="pill"
                    style={{ padding: '10px 24px' }}
                    onClick={() => {
                      setOrderSuccess(null);
                      setActiveModal(null);
                    }}
                  >
                    Close
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* AUTH MODAL */}
      {activeModal === 'auth' && (
        <div className="modal-overlay" onClick={() => setActiveModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{authMode === 'login' ? 'Board the Vessel' : 'Sign Articles'}</h2>
              <button className="modal-close" onClick={() => setActiveModal(null)}>✕</button>
            </div>
            {authError && <div style={{ color: 'red', marginBottom: 14 }}>{authError}</div>}
            <form onSubmit={handleAuth}>
              {authMode === 'register' && (
                <>
                  <div className="form-group">
                    <label>Sailor Name</label>
                    <input
                      type="text"
                      required
                      className="form-input"
                      value={authName}
                      onChange={(e) => setAuthName(e.target.value)}
                    />
                  </div>
                  <div className="form-group">
                    <label>Role</label>
                    <select
                      className="form-input"
                      value={authRole}
                      onChange={(e) => setAuthRole(e.target.value)}
                    >
                      <option value="user">Customer (Voyager)</option>
                      <option value="admin">Store Owner (Quartermaster)</option>
                    </select>
                  </div>
                </>
              )}
              <div className="form-group">
                <label>Email Address</label>
                <input
                  type="email"
                  required
                  className="form-input"
                  value={authEmail}
                  onChange={(e) => setAuthEmail(e.target.value)}
                />
              </div>
              <div className="form-group">
                <label>Secret Password</label>
                <input
                  type="password"
                  required
                  className="form-input"
                  value={authPassword}
                  onChange={(e) => setAuthPassword(e.target.value)}
                />
              </div>
              <button type="submit" className="pill" style={{ width: '100%', marginTop: 12, padding: '12px 0', textAlign: 'center' }}>
                {authMode === 'login' ? 'Enter Cabin' : 'Sign and Register'}
              </button>
            </form>
            <div style={{ marginTop: 18, textAlign: 'center', fontSize: 13 }}>
              {authMode === 'login' ? (
                <span>No parchment yet? <a href="#" style={{ textDecoration: 'underline' }} onClick={() => setAuthMode('register')}>Sign articles</a></span>
              ) : (
                <span>Already signed? <a href="#" style={{ textDecoration: 'underline' }} onClick={() => setAuthMode('login')}>Board here</a></span>
              )}
            </div>
          </div>
        </div>
      )}

      {/* STORE OWNER ADMIN MODAL (CAPTAIN'S DECK) */}
      {activeModal === 'admin' && (
        <div className="modal-overlay" onClick={() => setActiveModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 880 }}>
            <div className="modal-header">
              <h2>Captain's Deck & Inventory Manifest</h2>
              <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
                <button className="pill" style={{ padding: '6px 14px', fontSize: 11 }} onClick={handleLogout}>Abandon Ship</button>
                <button className="modal-close" onClick={() => setActiveModal(null)}>✕</button>
              </div>
            </div>

            {/* Store Stats Summary */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 20 }}>
              <div style={{ background: '#ece7d8', border: '2px solid var(--line)', padding: 14, borderRadius: 12, textAlign: 'center' }}>
                <div style={{ fontSize: 12, textTransform: 'uppercase', color: '#666' }}>Active Products</div>
                <div style={{ fontSize: 24, fontWeight: 'bold' }}>{adminProducts.length}</div>
              </div>
              <div style={{ background: '#ece7d8', border: '2px solid var(--line)', padding: 14, borderRadius: 12, textAlign: 'center' }}>
                <div style={{ fontSize: 12, textTransform: 'uppercase', color: '#666' }}>Items Sold</div>
                <div style={{ fontSize: 24, fontWeight: 'bold' }}>{totalItemsSold}</div>
              </div>
              <div style={{ background: '#ece7d8', border: '2px solid var(--line)', padding: 14, borderRadius: 12, textAlign: 'center' }}>
                <div style={{ fontSize: 12, textTransform: 'uppercase', color: '#666' }}>Total Revenue</div>
                <div style={{ fontSize: 24, fontWeight: 'bold', color: '#2d6a4f' }}>${totalRevenue.toFixed(2)}</div>
              </div>
            </div>

            {/* Create or Edit Product Form */}
            <div style={{ background: '#dcd7c7', padding: 20, borderRadius: 16, border: '2px solid var(--line)', marginBottom: 24 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h3>{editingProduct ? `Edit Cargo Item: ${editingProduct.name}` : 'Add New Item to Your Inventory'}</h3>
                {editingProduct && (
                  <button className="pill" style={{ padding: '4px 10px', fontSize: 11 }} onClick={cancelEditProduct}>
                    Cancel Edit
                  </button>
                )}
              </div>
              <form onSubmit={handleCreateProduct} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginTop: 12 }}>
                <input
                  type="text"
                  placeholder="Item Name (e.g. Spyglass)"
                  required
                  className="form-input"
                  value={newProdName}
                  onChange={(e) => setNewProdName(e.target.value)}
                />
                <input
                  type="number"
                  step="0.01"
                  placeholder="Price in Gold ($)"
                  required
                  className="form-input"
                  value={newProdPrice}
                  onChange={(e) => setNewProdPrice(e.target.value)}
                />
                <input
                  type="number"
                  placeholder="Stock Quantity"
                  required
                  className="form-input"
                  value={newProdStock}
                  onChange={(e) => setNewProdStock(e.target.value)}
                />
                <input
                  type="text"
                  placeholder="Description"
                  className="form-input"
                  value={newProdDesc}
                  onChange={(e) => setNewProdDesc(e.target.value)}
                />
                <button type="submit" className="pill" style={{ gridColumn: 'span 2', padding: '10px 0', textAlign: 'center', background: editingProduct ? '#d4edda' : '#ece7d8' }}>
                  {editingProduct ? 'Save Cargo Changes' : 'Commission Cargo Item'}
                </button>
              </form>
            </div>

            {/* Store Owner Products List */}
            <h3 style={{ marginBottom: 12 }}>Your Managed Products ({adminProducts.length})</h3>
            {adminProducts.length === 0 ? (
              <p style={{ marginBottom: 20 }}>No products listed under your name yet.</p>
            ) : (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 12, marginBottom: 24 }}>
                {adminProducts.map((p) => (
                  <div key={p._id} style={{ border: '2px solid var(--line)', padding: 12, borderRadius: 12, background: '#f4efe2' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <strong>{p.name}</strong>
                      <span>${p.price.toFixed(2)}</span>
                    </div>
                    <div style={{ fontSize: 12, color: '#555', margin: '6px 0' }}>Stock: {p.stock} units</div>
                    <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
                      <button
                        className="pill"
                        style={{ padding: '4px 10px', fontSize: 11, background: '#d4edda' }}
                        onClick={() => startEditProduct(p)}
                      >
                        Edit
                      </button>
                      <button
                        className="pill"
                        style={{ padding: '4px 10px', fontSize: 11, background: '#f8d7da' }}
                        onClick={() => handleDeleteProduct(p._id)}
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Manage Store Orders */}
            <h3 style={{ marginBottom: 12 }}>Orders For Your Trade Goods ({adminOrders.length})</h3>
            {adminOrders.length === 0 ? (
              <p>No customers have placed orders for your goods yet.</p>
            ) : (
              <div>
                {adminOrders.map((ord) => (
                  <div key={ord._id} style={{ border: '2px solid var(--line)', padding: 14, borderRadius: 12, marginBottom: 10, background: '#ece7d8' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <strong>Order #{ord._id.slice(0, 8)}</strong>
                      <span className="badge" style={{ background: ord.status === 'DELIVERED' ? '#d4edda' : '#fff3cd' }}>
                        {ord.status}
                      </span>
                    </div>
                    <div style={{ fontSize: 13, margin: '8px 0' }}>
                      {ord.items.map((it, idx) => (
                        <div key={idx}>• {it.productName} (Qty: {it.quantity} × ${it.price})</div>
                      ))}
                    </div>
                    <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
                      <button className="pill" style={{ padding: '6px 12px', fontSize: 11 }} onClick={() => handleUpdateOrderStatus(ord._id, 'SHIPPED')}>
                        Mark Shipped
                      </button>
                      <button className="pill" style={{ padding: '6px 12px', fontSize: 11 }} onClick={() => handleUpdateOrderStatus(ord._id, 'DELIVERED')}>
                        Mark Delivered
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
